import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import postcss from 'postcss'

const frontendRoot = join(fileURLToPath(new URL('..', import.meta.url)))
const generatedPath = join(frontendRoot, 'src', 'styles', 'generated', '_component-aliases.scss')
const requestedPrefix = process.argv[2]

if (!requestedPrefix || !/^[a-z0-9-]+$/.test(requestedPrefix)) {
  console.error('Usage: node scripts/style-migration-report.mjs <legacy-component-prefix>')
  console.error('Example: node scripts/style-migration-report.mjs components-payment-subscription-plan-card')
  process.exit(1)
}

const classPrefix = `.${requestedPrefix}__`
const root = postcss.parse(await readFile(generatedPath, 'utf8'), { from: generatedPath })
let ruleCount = 0

root.walkRules((rule) => {
  const matchingSelectors = rule.selectors.filter((selector) => selector.includes(classPrefix))
  if (matchingSelectors.length === 0) return

  ruleCount += 1
  const atRuleParents = []
  let parent = rule.parent
  while (parent && parent.type !== 'root') {
    if (parent.type === 'atrule') atRuleParents.unshift(`@${parent.name} ${parent.params}`.trim())
    parent = parent.parent
  }

  if (atRuleParents.length > 0) console.log(atRuleParents.join(' > '))
  console.log(`${matchingSelectors.join(',\n')} {`)
  for (const node of rule.nodes ?? []) console.log(`  ${node.toString()}`)
  console.log('}\n')
})

if (ruleCount === 0) {
  console.error(`No generated selectors found for ${requestedPrefix}.`)
  process.exit(2)
}

console.error(`Reported ${ruleCount} generated rule groups for ${requestedPrefix}.`)
