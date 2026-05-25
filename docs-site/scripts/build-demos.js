#!/usr/bin/env node
import { execSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const rootDir = path.resolve(__dirname, '..', '..')
const publicDir = path.resolve(__dirname, '..', 'public')

const targets = [
  { name: 'admin', srcDir: path.join(rootDir, 'web', 'admin') },
  { name: 'user',  srcDir: path.join(rootDir, 'web', 'user')  },
]

for (const t of targets) {
  console.log(`[demo] Building ${t.name}...`)
  execSync(`pnpm run build:demo --base=/${t.name}-demo/`, {
    cwd: t.srcDir,
    stdio: 'inherit',
  })

  const dist = path.join(t.srcDir, 'dist')
  if (!fs.existsSync(dist)) {
    throw new Error(`Build output not found: ${dist}`)
  }

  const out = path.join(publicDir, `${t.name}-demo`)
  fs.rmSync(out, { recursive: true, force: true })
  fs.mkdirSync(out, { recursive: true })
  copyDir(dist, out)
  console.log(`[demo] ${t.name} → ${out}`)
}

function copyDir(src, dest) {
  for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
    const s = path.join(src, entry.name)
    const d = path.join(dest, entry.name)
    if (entry.isDirectory()) {
      fs.mkdirSync(d, { recursive: true })
      copyDir(s, d)
    } else {
      fs.copyFileSync(s, d)
    }
  }
}
