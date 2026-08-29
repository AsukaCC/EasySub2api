import { readdir, readFile, writeFile } from 'node:fs/promises'
import { extname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = join(fileURLToPath(new URL('..', import.meta.url)))
const sourceRoot = join(frontendRoot, 'src')
const outputPath = join(frontendRoot, 'scripts', 'style-legacy-baseline.json')
const policyOutputPath = join(frontendRoot, 'scripts', 'style-policy-baseline.json')
const extensions = new Set(['.scss', '.vue', '.ts', '.tsx', '.js', '.jsx'])
const legacyClassPattern = /\b(?:components|views|features|composables|utils)-[a-z0-9-]+__[a-z0-9-]+\b/g
const numberedElementPattern = /__(?:panel|action|text|field|label|heading|description|icon|state|header|body|footer|main|section|navigation|link|router-link|pre|code)-\d+$/

async function collectFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true })
  const nested = await Promise.all(entries.map(async (entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return collectFiles(path)
    return extensions.has(extname(entry.name)) ? [path] : []
  }))
  return nested.flat()
}

const classes = new Set()
const policyClasses = new Set()
for (const path of await collectFiles(sourceRoot)) {
  const source = await readFile(path, 'utf8')
  for (const match of source.matchAll(legacyClassPattern)) classes.add(match[0])
  for (const match of source.matchAll(/\.([a-z_][a-z0-9_-]*(?:__[a-z0-9_-]+|--[a-z0-9_-]+))/gi)) {
    const className = match[1]
    const blockName = className.split(/__|--/, 1)[0]
    if (
      !classes.has(className) &&
      (className.length > 48 || blockName.length > 24 || numberedElementPattern.test(className))
    ) {
      policyClasses.add(className)
    }
  }
}

const sortedClasses = [...classes].sort()
await writeFile(outputPath, `${JSON.stringify(sortedClasses, null, 2)}\n`)
await writeFile(policyOutputPath, `${JSON.stringify([...policyClasses].sort(), null, 2)}\n`)
console.log(`Wrote ${sortedClasses.length} legacy classes to ${outputPath}.`)
console.log(`Wrote ${policyClasses.size} existing policy exceptions to ${policyOutputPath}.`)
