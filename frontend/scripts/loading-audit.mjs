import { readdir, readFile } from 'node:fs/promises'
import { extname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = fileURLToPath(new URL('..', import.meta.url))
const sourceRoot = join(frontendRoot, 'src')
const feedbackPath = join(sourceRoot, 'styles', '_feedback.scss')
const copiedSpinnerPath = 'M4 12a8 8 0 018-8V0C5.373'
const failures = []

async function collectFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true })
  const nested = await Promise.all(entries.map(async (entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return collectFiles(path)
    return extname(entry.name) === '.vue' ? [path] : []
  }))
  return nested.flat()
}

const feedback = await readFile(feedbackPath, 'utf8')
for (const name of ['spin', 'pulse', 'ping']) {
  if (!feedback.includes(`@keyframes ${name}`)) {
    failures.push(`src/styles/_feedback.scss [missing-keyframes] ${name}`)
  }
}

for (const path of await collectFiles(sourceRoot)) {
  const content = await readFile(path, 'utf8')
  if (!content.includes(copiedSpinnerPath)) continue
  const line = content.slice(0, content.indexOf(copiedSpinnerPath)).split(/\r?\n/).length
  failures.push(`${relative(frontendRoot, path)}:${line} [copied-spinner-svg] use LoadingSpinner`)
}

if (failures.length > 0) {
  console.error(`Loading audit failed with ${failures.length} violation(s):`)
  console.error(failures.join('\n'))
  process.exit(1)
}

console.log('Loading audit passed: shared keyframes exist and copied spinner SVGs are absent.')
