const fs = require('fs');
const path = require('path');

const dir = path.join(__dirname, 'src/app/components');
const files = fs.readdirSync(dir).filter(f => f.endsWith('.tsx'));

let changed = 0;

for (const file of files) {
  if (file === 'Login.tsx' || file === 'Register.tsx') continue;
  
  const filePath = path.join(dir, file);
  let content = fs.readFileSync(filePath, 'utf8');
  
  if (content.includes('<Leaf')) {
    // Replace <Leaf ... />
    content = content.replace(/<Leaf[^>]*\/>/g, '<img src={logoUrl} alt="Buen Vivir" className="w-8 h-8 object-contain" />');
    
    // Add import
    if (!content.includes("import logoUrl from './logo.png';")) {
      content = content.replace(/import .*?lucide-react.*?;\n/m, (match) => {
        return match + "import logoUrl from './logo.png';\n";
      });
    }
    
    fs.writeFileSync(filePath, content, 'utf8');
    console.log('Fixed ' + file);
    changed++;
  }
}

console.log('Total files changed: ' + changed);
