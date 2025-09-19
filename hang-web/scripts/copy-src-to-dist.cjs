const fs = require('fs');
const path = require('path');

const srcDir = path.resolve(__dirname, '..', 'src');
const distDir = path.resolve(__dirname, '..', 'dist');

function copyRecursive(src, dest) {
  const stat = fs.statSync(src);
  if (stat.isDirectory()) {
    if (!fs.existsSync(dest)) fs.mkdirSync(dest, { recursive: true });
    for (const entry of fs.readdirSync(src)) {
      copyRecursive(path.join(src, entry), path.join(dest, entry));
    }
  } else {
  if (/\.tsx?$/.test(src) && !/(?:\.test|\.spec)\.tsx?$/.test(src) && path.basename(src) !== 'test.ts') {
      const destDir = path.dirname(dest);
      if (!fs.existsSync(destDir)) fs.mkdirSync(destDir, { recursive: true });
      fs.copyFileSync(src, dest);
    }
  }
}

if (!fs.existsSync(srcDir)) {
  console.error('src dir not found:', srcDir);
  process.exit(1);
}

copyRecursive(srcDir, distDir);
console.log('copied src to dist');
