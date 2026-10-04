// Import the pinned public bank and regenerate golden scores using upstream JS.
import { readFileSync, writeFileSync } from 'node:fs'
import { execFileSync } from 'node:child_process'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const revision = 'd4131b30243dfa05e70180b5eedde742103f1d73'
const source = resolve(process.argv[2] || '')
if (execFileSync('git', ['-C', source, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim() !== revision) {
  throw new Error(`Expected ModelTrace revision ${revision}`)
}
const target = fileURLToPath(new URL('../../backend/internal/pkg/modeltrace/', import.meta.url))
const bankText = readFileSync(resolve(source, 'data/unified_bank.json'), 'utf8')
const bank = JSON.parse(bankText)
const scorerText = readFileSync(resolve(source, 'static/fingerprint-core.js'), 'utf8')
const scorer = await import(`data:text/javascript;base64,${Buffer.from(scorerText).toString('base64')}`)
const cases = JSON.parse(readFileSync(resolve(target, 'testdata/parity.json'), 'utf8'))
for (const test of cases) {
  const result = scorer.analyzeGlobalOutputs(test.outputs.map(text => ({ text })), bank)
  test.expected = result.results.map(item => ({ model: item.model, family: item.family, probability: item.probability }))
}
writeFileSync(resolve(target, 'unified_bank.json'), bankText)
writeFileSync(resolve(target, 'testdata/parity.json'), `${JSON.stringify(cases, null, 2)}\n`)
console.log(`Imported ${bank.models.length} candidates at ${revision}`)
