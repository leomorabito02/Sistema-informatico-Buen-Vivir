import re

with open('.github/workflows/pipeline.yml', 'r') as f:
    content = f.read()

# 1. CI-RXRESUME AND CI-REACTIVE-RESUME
ci_rxresume_block = """
  ci-rxresume:
    name: CI — RxResume Service (Go)
    runs-on: ubuntu-latest
    timeout-minutes: 15

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
          cache-dependency-path: rxresume/go.sum

      - name: Download Go modules
        working-directory: rxresume
        run: go mod download

      - name: Run Go Unit Tests
        working-directory: rxresume
        run: go test -v -coverprofile=coverage.out ./...
        
      - name: Go Security Scan (gosec)
        working-directory: rxresume
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec -fmt sarif -out gosec-results.sarif ./...
        continue-on-error: true

      - name: Docker build (rxresume)
        run: docker build -t rxresume:${{ github.sha }} -f rxresume/Dockerfile ./rxresume

      - name: Authenticate to Google Cloud
        uses: google-github-actions/auth@v2
        with:
          credentials_json: '${{ secrets.GCP_CREDENTIALS }}'

      - name: Setup Google Cloud SDK
        uses: google-github-actions/setup-gcloud@v2

      - name: Configure Docker for Artifact Registry
        run: gcloud auth configure-docker ${{ secrets.ARTIFACT_REGISTRY_HOST }} --quiet

      - name: Push rxresume image to Artifact Registry
        run: |
          IMAGE_URI="${{ secrets.ARTIFACT_REGISTRY_HOST }}/${{ secrets.GCP_PROJECT_ID }}/${{ secrets.ARTIFACT_REGISTRY_REPO }}/rxresume:${{ github.sha }}"
          docker tag rxresume:${{ github.sha }} "$IMAGE_URI"
          docker push "$IMAGE_URI"

  ci-reactive-resume:
    name: CI — Reactive Resume (Node)
    runs-on: ubuntu-latest
    timeout-minutes: 20

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}

      - name: Setup pnpm
        uses: pnpm/action-setup@v4
        with:
          version: 9
          run_install: false

      - name: Get pnpm store directory
        shell: bash
        run: |
          echo "STORE_PATH=$(pnpm store path --silent)" >> $GITHUB_ENV

      - name: Setup pnpm cache
        uses: actions/cache@v4
        with:
          path: ${{ env.STORE_PATH }}
          key: ${{ runner.os }}-pnpm-store-${{ hashFiles('rxresume/reactive-resume/pnpm-lock.yaml') }}
          restore-keys: |
            ${{ runner.os }}-pnpm-store-

      - name: Install dependencies
        working-directory: rxresume/reactive-resume
        run: pnpm install --frozen-lockfile

      - name: Build Reactive Resume
        working-directory: rxresume/reactive-resume
        run: pnpm build

      - name: Test Reactive Resume
        working-directory: rxresume/reactive-resume
        run: pnpm test

      - name: Docker build (reactive-resume)
        run: docker build -t reactive-resume:${{ github.sha }} -f rxresume/reactive-resume/Dockerfile ./rxresume/reactive-resume

      - name: Authenticate to Google Cloud
        uses: google-github-actions/auth@v2
        with:
          credentials_json: '${{ secrets.GCP_CREDENTIALS }}'

      - name: Setup Google Cloud SDK
        uses: google-github-actions/setup-gcloud@v2

      - name: Configure Docker for Artifact Registry
        run: gcloud auth configure-docker ${{ secrets.ARTIFACT_REGISTRY_HOST }} --quiet

      - name: Push reactive-resume image to Artifact Registry
        run: |
          IMAGE_URI="${{ secrets.ARTIFACT_REGISTRY_HOST }}/${{ secrets.GCP_PROJECT_ID }}/${{ secrets.ARTIFACT_REGISTRY_REPO }}/reactive-resume:${{ github.sha }}"
          docker tag reactive-resume:${{ github.sha }} "$IMAGE_URI"
          docker push "$IMAGE_URI"
"""

# Insert before ci-api-gateway
content = re.sub(r'(  ci-api-gateway:)', ci_rxresume_block + r'\n\1', content)

# 2. MUTATION-RXRESUME
mutation_rxresume_block = """
  mutation-rxresume:
    name: Mutation — RxResume Service (gremlins)
    runs-on: ubuntu-latest
    needs: [ci-rxresume]
    if: false # Disable for now unless explicitly needed
    timeout-minutes: 30

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
          cache-dependency-path: rxresume/go.sum

      - name: Download Go modules
        working-directory: rxresume
        run: go mod download

      - name: Install gremlins
        run: go install github.com/go-gremlins/gremlins/cmd/gremlins@latest

      - name: Run mutation tests (rxresume)
        working-directory: rxresume
        run: |
          gremlins unleash ./... || echo "Gremlins found survivors"
"""

# Insert after mutation-education
content = re.sub(r'(  cd-setup:)', mutation_rxresume_block + r'\n\1', content)

# 3. CD-RXRESUME AND CD-REACTIVE-RESUME
cd_rxresume_block = """
  cd-rxresume:
    name: CD — RxResume Service (Cloud Run)
    runs-on: ubuntu-latest
    needs: [cd-setup]
    environment: ${{ needs.cd-setup.outputs.env_name }}
    timeout-minutes: 15
    env:
      APP_ENV: ${{ needs.cd-setup.outputs.env_name == 'PROD' && 'production' || 'development' }}
    outputs:
      service_url: ${{ steps.deploy.outputs.url }}

    steps:
      - name: Authenticate to Google Cloud
        uses: google-github-actions/auth@v2
        with:
          credentials_json: '${{ secrets.GCP_CREDENTIALS }}'

      - name: Verify Image exists
        id: pre-deploy
        run: |
          IMAGE="${{ secrets.ARTIFACT_REGISTRY_HOST }}/${{ secrets.GCP_PROJECT_ID }}/${{ secrets.ARTIFACT_REGISTRY_REPO }}/rxresume:${{ github.sha }}"
          gcloud container images describe "$IMAGE" || (echo "::error::Image $IMAGE not found in registry" && exit 1)

      - name: Deploy RxResume to Cloud Run
        id: deploy
        uses: google-github-actions/deploy-cloudrun@v2
        with:
          service: 'rxresume-${{ needs.cd-setup.outputs.service_suffix }}'
          region: '${{ secrets.GCP_REGION }}'
          image: '${{ secrets.ARTIFACT_REGISTRY_HOST }}/${{ secrets.GCP_PROJECT_ID }}/${{ secrets.ARTIFACT_REGISTRY_REPO }}/rxresume:${{ github.sha }}'
          flags: '--no-allow-unauthenticated --min-instances=0 --max-instances=1 --cpu=1 --memory=1Gi --service-account=${{ secrets.GCP_SERVICE_ACCOUNT }}'

  cd-reactive-resume:
    name: CD — Reactive Resume (Cloud Run)
    runs-on: ubuntu-latest
    needs: [cd-setup]
    environment: ${{ needs.cd-setup.outputs.env_name }}
    timeout-minutes: 15
    outputs:
      service_url: ${{ steps.deploy.outputs.url }}

    steps:
      - name: Authenticate to Google Cloud
        uses: google-github-actions/auth@v2
        with:
          credentials_json: '${{ secrets.GCP_CREDENTIALS }}'

      - name: Verify Image exists
        id: pre-deploy
        run: |
          IMAGE="${{ secrets.ARTIFACT_REGISTRY_HOST }}/${{ secrets.GCP_PROJECT_ID }}/${{ secrets.ARTIFACT_REGISTRY_REPO }}/reactive-resume:${{ github.sha }}"
          gcloud container images describe "$IMAGE" || (echo "::error::Image $IMAGE not found in registry" && exit 1)

      - name: Deploy Reactive Resume to Cloud Run
        id: deploy
        uses: google-github-actions/deploy-cloudrun@v2
        with:
          service: 'reactive-resume-${{ needs.cd-setup.outputs.service_suffix }}'
          region: '${{ secrets.GCP_REGION }}'
          image: '${{ secrets.ARTIFACT_REGISTRY_HOST }}/${{ secrets.GCP_PROJECT_ID }}/${{ secrets.ARTIFACT_REGISTRY_REPO }}/reactive-resume:${{ github.sha }}'
          flags: '--allow-unauthenticated --min-instances=0 --max-instances=1 --cpu=1 --memory=1Gi --service-account=${{ secrets.GCP_SERVICE_ACCOUNT }}'
"""

# Insert before cd-api-gateway
content = re.sub(r'(  cd-api-gateway:)', cd_rxresume_block + r'\n\1', content)

# 4. Update dependencies
# 4.1 Update ci-gate
content = re.sub(
    r'(needs: \[.*ci-education.*?)(, ci-api-gateway\])',
    r'\1, ci-rxresume, ci-reactive-resume\2',
    content
)

# 4.2 Update cd-api-gateway needs
content = re.sub(
    r'(cd-api-gateway:[\s\S]*?needs: \[cd-setup, cd-auth, cd-education)\]',
    r'\1, cd-rxresume, cd-reactive-resume]',
    content
)

# 4.3 Update cd-api-gateway environment variables / urls if needed.
# Since rxresume is used in the gateway, we need to pass its URL to the gateway config build.
rxresume_gateway_env = """
          export RXRESUME_SERVICE_URL=$(gcloud run services describe "rxresume-$SUFFIX" --region="${{ secrets.GCP_REGION }}" --format='value(status.url)')
          export REACTIVE_RESUME_URL=$(gcloud run services describe "reactive-resume-$SUFFIX" --region="${{ secrets.GCP_REGION }}" --format='value(status.url)')
"""
content = re.sub(
    r'(export EDUCATION_SERVICE_URL=.*?\n)',
    r'\1' + rxresume_gateway_env,
    content
)

# 4.4 Update cd-smoke needs
content = re.sub(
    r'(cd-smoke:[\s\S]*?needs: \[cd-setup, cd-auth, cd-education)(, cd-api-gateway)',
    r'\1, cd-rxresume, cd-reactive-resume\2',
    content
)

# 4.5 Update cd-rollback needs
content = re.sub(
    r'(cd-rollback:[\s\S]*?needs: \[cd-setup, cd-auth, cd-education)',
    r'\1, cd-rxresume, cd-reactive-resume',
    content
)

# 4.6 Update cd-rollback delete statements
rollback_statements = """
          gcloud run services delete "rxresume-$SUFFIX" --region="${{ secrets.GCP_REGION }}" --quiet || true
          gcloud run services delete "reactive-resume-$SUFFIX" --region="${{ secrets.GCP_REGION }}" --quiet || true
"""
content = re.sub(
    r'(gcloud run services delete "education-\$SUFFIX" --region="\$\{\{ secrets\.GCP_REGION \}\}" --quiet \|\| true\n)',
    r'\1' + rollback_statements,
    content
)

# Update cd-dast-zap needs
content = re.sub(
    r'(cd-dast-zap:[\s\S]*?needs: \[cd-setup, cd-smoke, cd-auth, cd-education)(, cd-api-gateway\])',
    r'\1, cd-rxresume, cd-reactive-resume\2',
    content
)


with open('.github/workflows/pipeline.yml', 'w') as f:
    f.write(content)
