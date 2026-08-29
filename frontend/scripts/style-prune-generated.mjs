import { readFile, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import postcss from 'postcss'

const frontendRoot = join(fileURLToPath(new URL('..', import.meta.url)))
const generatedPath = join(frontendRoot, 'src', 'styles', 'generated', '_component-aliases.scss')
const requestedPrefix = process.argv[2]

if (!requestedPrefix || !/^[a-z0-9-]+$/.test(requestedPrefix)) {
  console.error('Usage: node scripts/style-prune-generated.mjs <legacy-component-prefix>')
  process.exit(1)
}

const classPrefix = `.${requestedPrefix}__`
const root = postcss.parse(await readFile(generatedPath, 'utf8'), { from: generatedPath })
let removedSelectors = 0

root.walkRules((rule) => {
  const retainedSelectors = rule.selectors.filter((selector) => {
    if (!selector.includes(classPrefix)) return true
    removedSelectors += 1
    return false
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

if (removedSelectors === 0) {
  console.error(`No generated selectors found for ${requestedPrefix}.`)
  process.exit(2)
}

await writeFile(generatedPath, root.toString())
console.log(`Removed ${removedSelectors} selectors for ${requestedPrefix}.`)
