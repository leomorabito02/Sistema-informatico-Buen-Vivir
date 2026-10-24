# Guía de los scripts y resumen para Leonardo

Este documento explica, sin tecnicismos, qué hace cada script que armamos para
el TP, dónde quedan los resultados de SonarQube y cómo contarle a Leonardo lo
que se hizo.

---

## 1. ¿Dónde quedan escritos los resultados de SonarQube?

En **tres lugares**, según desde dónde se haga el análisis:

| Dónde | Qué guarda | ¿Dura? |
|---|---|---|
| **SonarQube de tu PC** (http://localhost:9000) | Todo: los números, cada problema encontrado con el archivo y la línea, y la evolución en el tiempo | Sí, mientras no borres la carpeta `data` de SonarQube |
| **`docs/calidad/reportes/`** en el repo | `historial.csv`: una fila por análisis con fecha, commit, resultado del quality gate y unas 30 métricas. Además, un `.json` por análisis con la respuesta completa de SonarQube, como respaldo | Sí, porque se sube a GitHub con cada commit |
| **GitHub Actions** (pipeline `calidad-tp.yml`) | En cada corrida: una tabla resumen en la pantalla de la corrida y un archivo descargable `calidad-education-<commit>` con los reportes | El resumen queda siempre. El archivo descargable se borra a los 30 días |

**En simple:** SonarQube muestra los resultados en pantalla, pero nosotros
además los **copiamos a un archivo** (`historial.csv`) que queda en el repo.
Así los números del informe se pueden comprobar: cada fila dice de qué commit
salió. La consigna lo exige: "los números del informe tienen que coincidir con
los de la herramienta".

El SonarQube del pipeline se borra al terminar cada corrida. Lo que queda de
él es el resumen y el archivo descargable.

---

## 2. Los scripts, uno por uno

Un **script** es una receta: una lista de pasos escrita para que la
computadora los haga sola, siempre igual. Todos se corren desde la carpeta
principal del repo, en **Git Bash**.

### Carpeta `scripts/calidad/`: análisis con SonarQube

| Script | Qué hace, en simple | Cuándo se usa |
|---|---|---|
| `configurar_sonar.sh` | Prepara SonarQube: crea el proyecto "Buen Vivir - Educación" y el quality gate "TP-Calidad-N3" con las 6 reglas del Nivel 3. Si ya existen, los deja igual | Una vez en tu PC. El pipeline lo corre en cada corrida, porque su SonarQube arranca vacío |
| `linea_base.sh` | Saca una "foto" de cómo era Educación el 1/8 (antes de agosto) y la registra con fecha 31/07. Sin esa foto, SonarQube no sabe qué código es "nuevo" | Una sola vez, en un proyecto recién creado. El pipeline también lo corre cada vez |
| `analizar.sh` | El principal: corre las pruebas de Educación, mide la cobertura, se la pasa a SonarQube, espera el veredicto del quality gate y anota todo en el historial | Cada vez que quieras medir. Si el gate da rojo, termina con error |
| `exportar_metricas.py` | Le pide los números a SonarQube y escribe la fila en `historial.csv` y el `.json` | No se usa a mano: lo llama `analizar.sh` al final |

Comando para medir en tu PC (con SonarQube y el emulador de Firestore
prendidos):

```bash
FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 \
SONAR_SCANNER="C:/Users/sofic/tools/sonar-scanner-8.1.0.6389-windows-x64/bin/sonar-scanner.bat" \
bash scripts/calidad/analizar.sh education
```

### Carpeta `scripts/calidad/entorno/`: el entorno de prueba

Estos scripts acompañan a `docker-compose.calidad.yml`, que prende una
**copia de Educación aparte**, con datos inventados, para probar sin tocar
nada real.

| Script | Qué hace, en simple | Cuándo se usa |
|---|---|---|
| `preparar_env.sh` | Arma el archivo de configuración del entorno (`.env.calidad`). Copia de `education/.env` solo las claves de Contentful, y solo si apuntan a `testing`. Además inventa una clave nueva para firmar los usuarios de prueba. No muestra ninguna clave en pantalla, y el archivo no se sube a GitHub | Una sola vez, antes del primer `docker compose up` |
| `cargar_por_api.py` | Carga los datos de prueba **usando la aplicación**, como lo haría la gente: un admin crea 2 actividades por curso y les pone habilidades, inscribe a 8 participantes de prueba y 5 de ellos entregan la primera actividad. Al entregar, Educación registra la asistencia y otorga las habilidades sola. Cada dato pasa por las mismas reglas que uno real. Los cursos son los reales de Contentful `testing`: los lee, no los cambia | Cada vez que prendés el entorno, porque arranca vacío. Se puede repetir: lo que ya existe lo saltea |
| `auth_falso.py` | Un auth-service de prueba. Educación le pregunta a auth-service si un participante existe y está aprobado, y "quién soy" cuando alguien entrega. Este responde eso mismo para 8 participantes de prueba (`participante-prueba-01` a `08`), con el mismo formato que el real | Lo prende solo `docker compose`; no hay que hacer nada |
| `token_prueba.py` | Fabrica un "pase" (token) para entrar como admin, docente, voluntario o participante de prueba. Dura 60 minutos y solo sirve en este entorno | Cuando quieras probar una pantalla o un pedido con un rol concreto |
| `README.md` | Las instrucciones de uso, paso a paso | Para leer |

**Ejemplo para entenderlo:** el entorno es como una maqueta de Educación. El
`docker compose up` arma la maqueta vacía, `cargar_por_api.py` la usa como
lo haría la gente (crea actividades, inscribe, entrega) y `token_prueba.py`
te da la credencial para entrar con el rol que quieras.

---

## 3. Cómo explicarle a Leonardo lo que hiciste

### La idea en una frase

> "Armé la parte de **medir la calidad del código**: SonarQube, el quality
> gate del Nivel 3, un pipeline que solo hace el análisis estático y un
> entorno de prueba con datos inventados. **Tus pruebas son las que hacen
> subir la cobertura**, y mi parte la mide."

### Los puntos que le tenés que contar

1. **El TP es solo sobre Educación.** CVs (rxresume, cv-renderer) quedó afuera
   (decisión D-19).
2. **SonarQube ya mide Educación.** Primera medición, el 17/10, antes de
   sus pruebas nuevas del 18/10:
   - 8.088 líneas de código;
   - 0 bugs y 0 vulnerabilidades;
   - 33 problemas de mantenibilidad;
   - **80,4 % de cobertura** (apenas arriba de la meta de 80 %);
   - 0,2 % de duplicación.

   **Segunda medición, el 18/10, con su commit `42414bc`:** la cobertura
   subió a **92,6 %** (12,2 puntos más) y las pruebas pasaron de 466 a 886.
   Es evidencia para los dos. Aviso: SonarQube marcó 3 funciones de sus
   pruebas como demasiado complejas (`contentful_client_helpers_test.go:52`,
   `material_repository_test.go:183` y `material_repository_upload_test.go:131`).
3. **El quality gate "TP-Calidad-N3"** frena si hay bugs críticos, si la
   cobertura baja de 80 % (en todo el código o en el nuevo) o si la
   duplicación pasa el 5 %. "Código nuevo" es lo hecho desde el 1/8.
4. **El pipeline `calidad-tp.yml` es solo de análisis estático (eje 4).**
   Corre las pruebas que ya existen, incluidas las suyas, para medir la
   cobertura, pero **no agrega pruebas** ni reemplaza su pipeline. El
   pipeline de la tesis (`pipeline.yml`) no se tocó.
5. **El entorno de calidad** (`docker-compose.calidad.yml`) le puede servir
   para sus pruebas de API, smoke o E2E: es Educación con datos inventados
   y Contentful `testing`, en el puerto 18210. Las instrucciones están en
   `scripts/calidad/entorno/README.md`.
6. **Lo que necesitás de él (el "contrato"):** que sus pruebas sigan
   dejando la cobertura en `education/coverage.out`, el formato estándar de
   Go. Si después hace pruebas de carga con k6, que guarde el resumen en
   JSON, así se puede analizar.
7. **Dos avisos (hallazgos):**
   - **H-01:** las pruebas de Educación cambian la fecha de dos reportes
     versionados (`reports/database/...` y `reports/pact/...`). Su commit
     también los trae modificados. Conviene que esas pruebas escriban en una
     carpeta que no se suba, o revertirlos antes de cada commit.
   - **H-04:** el reporte de k6 de `reports/performance/` lo inventa un
     script con valores fijos, así que no sirve como evidencia.
8. **Todo lo decidido está en `docs/calidad/decisiones.md`**, con fecha y
   quién lo decidió. La visión general está en `docs/calidad/00-vision-general.md`.

### Mensaje listo para mandarle

> Leo, te cuento lo que hice del TP de Calidad (está todo en la rama Calidad,
> commit "probando sonarqube"):
>
> - El TP queda solo sobre Educación.
> - Configuré SonarQube con un quality gate del Nivel 3 (bugs críticos,
>   cobertura ≥ 80 %, duplicación ≤ 5 %). La primera medición, del 17/10,
>   antes de tus tests nuevos, dio 80,4 % de cobertura y 0 bugs. Con tus
>   tests (commit 42414bc) subió a 92,6 %, y las pruebas pasaron de 466
>   a 886. SonarQube marcó 3 funciones de tus tests como muy complejas:
>   te paso la lista.
> - Hice un pipeline aparte, `calidad-tp.yml`, que solo hace el análisis de
>   SonarQube. Usa tus tests para medir la cobertura, pero no toca tu
>   pipeline ni agrega pruebas.
> - Hay un entorno de prueba con datos inventados (`docker-compose.calidad.yml`,
>   puerto 18210). Te puede servir para smoke o E2E. Instrucciones en
>   `scripts/calidad/entorno/README.md`.
> - Lo único que necesito de tu lado: que la cobertura siga saliendo en
>   `education/coverage.out`, y si hacés k6, el resumen en JSON.
> - Ojo: los tests cambian la fecha de `reports/database` y `reports/pact`
>   en cada corrida, y el reporte k6 de `reports/performance` es inventado
>   por un script.
>
> Las decisiones están en `docs/calidad/decisiones.md`.
