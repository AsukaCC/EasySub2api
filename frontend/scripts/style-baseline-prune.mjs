import { readdir, readFile, writeFile } from 'node:fs/promises'
import { extname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = join(fileURLToPath(new URL('..', import.meta.url)))
const sourceRoot = join(frontendRoot, 'src')
const scriptsRoot = join(frontendRoot, 'scripts')
const extensions = new Set(['.scss', '.vue', '.ts', '.tsx', '.js', '.jsx'])
const legacyClassPattern = /\b(?:components|views|features|composables|utils)-[a-z0-9-]+__[a-z0-9-]+\b/g
const classSelectorPattern = /\.([a-z_][a-z0-9_-]*(?:__[a-z0-9_-]+|--[a-z0-9_-]+))/gi

async function collectFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true })
  const nested = await Promise.all(entries.map(async (entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return collectFiles(path)
    return extensions.has(extname(entry.name)) ? [path] : []
  }))
  return nested.flat()
}

const legacyClasses = new Set()
const ownedClasses = new Set()
for (const path of await collectFiles(sourceRoot)) {
  const source = await readFile(path, 'utf8')
  for (const match of source.matchAll(legacyClassPattern)) legacyClasses.add(match[0])
  for (const match of source.matchAll(classSelectorPattern)) ownedClasses.add(match[1])
}

async function pruneBaseline(fileName, encounteredClasses) {
  const path = join(scriptsRoot, fileName)
  const existing = JSON.parse(await readFile(path, 'utf8'))
  const retained = existing.filter((className) => encounteredClasses.has(className))
  await writeFile(path, `${JSON.stringify(retained, null, 2)}\n`)
  console.log(`${fileName}: removed ${existing.length - retained.length}, retained ${retained.length}.`)
}

await pruneBaseline('style-legacy-baseline.json', legacyClasses)
await pruneBaseline('style-policy-baseline.json', ownedClasses)
