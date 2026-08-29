import { readdir, readFile, writeFile } from 'node:fs/promises'
import { extname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import postcss from 'postcss'

const frontendRoot = join(fileURLToPath(new URL('..', import.meta.url)))
const sourceRoot = join(frontendRoot, 'src')
const generatedPath = join(sourceRoot, 'styles', 'generated', '_component-aliases.scss')
const extensions = new Set(['.scss', '.vue', '.ts', '.tsx', '.js', '.jsx'])
const legacyClassPattern = /\b(?:components|views|features|composables|utils)-[a-z0-9-]+__[a-z0-9-]+\b/g

async function collectFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true })
  const nested = await Promise.all(entries.map(async (entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return collectFiles(path)
    return extensions.has(extname(entry.name)) ? [path] : []
  }))
  return nested.flat()
}

const referencedClasses = new Set()
for (const path of await collectFiles(sourceRoot)) {
  if (resolve(path) === resolve(generatedPath)) continue
  const source = await readFile(path, 'utf8')
  for (const match of source.matchAll(legacyClassPattern)) referencedClasses.add(match[0])
}

const root = postcss.parse(await readFile(generatedPath, 'utf8'), { from: generatedPath })
let removedSelectors = 0

root.walkRules((rule) => {
  const retainedSelectors = rule.selectors.filter((selector) => {
    const selectorClasses = [...selector.matchAll(legacyClassPattern)].map((match) => match[0])
    const isDead = selectorClasses.length > 0 && selectorClasses.some((className) => !referencedClasses.has(className))
    if (isDead) removedSelectors += 1
    return !isDead
  })

  if (retainedSelectors.length === 0) rule.remove()
  else if (retainedSelectors.length !== rule.selectors.length) rule.selectors = retainedSelectors
})

let removedEmptyAtRule = true
while (removedEmptyAtRule) {
  removedEmptyAtRule = false
  root.walkAtRules((atRule) => {
    if (atRule.nodes && atRule.nodes.length === 0) {
      atRule.remove()
      removedEmptyAtRule = true
    }
  })
}

await writeFile(generatedPath, root.toString())
console.log(`Removed ${removedSelectors} unreferenced generated selectors.`)
