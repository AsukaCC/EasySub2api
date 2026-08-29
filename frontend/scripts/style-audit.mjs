import { readdir, readFile } from 'node:fs/promises'
import { extname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = join(fileURLToPath(new URL('..', import.meta.url)))
const sourceRoot = join(frontendRoot, 'src')
const scriptsRoot = join(frontendRoot, 'scripts')
const failures = []
const warnings = []
const legacyClassPattern = /\b(?:components|views|features|composables|utils)-[a-z0-9-]+__[a-z0-9-]+\b/g
const numberedElementPattern = /__(?:panel|action|text|field|label|heading|description|icon|state|header|body|footer|main|section|navigation|link|router-link|pre|code)-\d+$/

async function collectFiles(directory, extensions) {
  const entries = await readdir(directory, { withFileTypes: true })
  const nested = await Promise.all(entries.map(async (entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return collectFiles(path, extensions)
    return extensions.includes(extname(entry.name)) ? [path] : []
  }))
  return nested.flat()
}

function report(path, rule, match) {
  const before = match.input.slice(0, match.index)
  const line = before.split(/\r?\n/).length
  failures.push(`${relative(frontendRoot, path)}:${line} [${rule}] ${match[0].trim()}`)
}

const styleFiles = await collectFiles(sourceRoot, ['.scss', '.vue'])
for (const path of styleFiles) {
  if (path.includes(`${join('', '__tests__')}`)) continue
  const content = await readFile(path, 'utf8')
  const isTokenFile = path.endsWith(`${join('styles', '_tokens.scss')}`)

  if (!isTokenFile) {
    for (const match of content.matchAll(/font-size\s*:\s*-?(?:\d*\.)?\d+(?:px|rem)\b/gi)) {
      report(path, 'raw-font-size', match)
    }
  }

  for (const match of content.matchAll(/letter-spacing\s*:\s*([^;}]+)/gi)) {
    if (match[1].trim() !== '0') report(path, 'letter-spacing', match)
  }

  for (const match of content.matchAll(/:hover[^{}]*\{[^{}]*background(?:-color)?\s*:\s*(?:#(?:fff(?:fff)?|000(?:000)?)\b|white\b|black\b|var\(--color-(?:primary|primary-hover|success|danger|warning)\))/gis)) {
    report(path, 'opaque-hover', match)
  }
}

const sourceFiles = await collectFiles(sourceRoot, ['.scss', '.vue', '.ts', '.tsx', '.js', '.jsx'])
const baselinePath = join(scriptsRoot, 'style-legacy-baseline.json')
const policyBaselinePath = join(scriptsRoot, 'style-policy-baseline.json')
const classExceptionsPath = join(scriptsRoot, 'style-class-exceptions.json')
const baseline = new Set(JSON.parse(await readFile(baselinePath, 'utf8')))
const policyBaseline = new Set(JSON.parse(await readFile(policyBaselinePath, 'utf8')))
const classExceptions = new Set(JSON.parse(await readFile(classExceptionsPath, 'utf8')))
const encounteredLegacyClasses = new Set()
const encounteredOwnedClasses = new Set()

for (const path of sourceFiles) {
  const content = await readFile(path, 'utf8')

  for (const match of content.matchAll(legacyClassPattern)) {
    encounteredLegacyClasses.add(match[0])
    if (!baseline.has(match[0])) report(path, 'new-legacy-class', match)
  }

  for (const match of content.matchAll(/\.([a-z_][a-z0-9_-]*(?:__[a-z0-9_-]+|--[a-z0-9_-]+))/gi)) {
    const className = match[1]
    encounteredOwnedClasses.add(className)
    if (baseline.has(className) || policyBaseline.has(className) || classExceptions.has(className)) continue

    const blockName = className.split(/__|--/, 1)[0]
    if (className.length > 48) report(path, 'class-name-too-long', match)
    if (blockName.length > 24) report(path, 'class-block-too-long', match)
    if (numberedElementPattern.test(className)) report(path, 'numbered-class-element', match)
  }
}

for (const className of baseline) {
  if (!encounteredLegacyClasses.has(className)) {
    failures.push(`scripts/style-legacy-baseline.json [stale-legacy-class] ${className}`)
  }
}

for (const className of policyBaseline) {
  if (!encounteredOwnedClasses.has(className)) {
    failures.push(`scripts/style-policy-baseline.json [stale-policy-class] ${className}`)
  }
}

for (const className of classExceptions) {
  if (!encounteredOwnedClasses.has(className)) {
    warnings.push(`scripts/style-class-exceptions.json [unused-class-exception] ${className}`)
  }
}

const tokenPath = join(sourceRoot, 'styles', '_tokens.scss')
const tokenSource = await readFile(tokenPath, 'utf8')
const darkSource = tokenSource.match(/\.dark\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
const pairedThemeTokens = [
  '--color-text-primary',
  '--color-text-secondary',
  '--color-text-tertiary',
  '--color-text-quaternary',
  '--color-text-muted',
  '--glass-bg-interactive',
  '--glass-field-bg',
  '--glass-saturate',
  '--glass-saturate-hover',
  '--glass-tint-brand',
  '--glass-tint-success',
  '--glass-tint-warning',
  '--glass-tint-danger',
  '--glass-layer-inset-bg',
  '--glass-layer-inset-blur',
  '--glass-layer-inset-blur-hover',
  '--glass-layer-content-bg',
  '--glass-layer-content-blur',
  '--glass-layer-content-blur-hover',
  '--glass-layer-shell-bg',
  '--glass-layer-shell-blur',
  '--glass-layer-shell-blur-hover',
  '--glass-layer-floating-bg',
  '--glass-layer-floating-blur',
  '--glass-layer-modal-bg',
  '--glass-layer-modal-blur',
  '--glass-layer-scrim-bg',
  '--glass-layer-scrim-blur',
]
for (const token of pairedThemeTokens) {
  if (!darkSource.includes(`${token}:`)) failures.push(`src/styles/_tokens.scss [missing-dark-token] ${token}`)
}

const canonicalLayerMappings = [
  ['src/styles/glass.scss', '@include glass-surface(floating'],
  ['src/styles/glass.scss', '@include glass-surface(modal'],
  ['src/components/common/Select.vue', 'var(--glass-layer-floating-bg)'],
  ['src/styles/_feedback.scss', 'var(--glass-layer-scrim-bg)'],
]
for (const [relativePath, requiredSource] of canonicalLayerMappings) {
  const source = await readFile(join(frontendRoot, relativePath), 'utf8')
  if (!source.includes(requiredSource)) {
    failures.push(`${relativePath} [missing-glass-layer-mapping] ${requiredSource}`)
  }
}

if (failures.length > 0) {
  console.error(`Style audit failed with ${failures.length} violation(s):`)
  console.error(failures.join('\n'))
  process.exit(1)
}

if (warnings.length > 0) console.warn(warnings.join('\n'))
console.log(
  `Style audit passed (${styleFiles.length} style-bearing files, ${sourceFiles.length} source files, ` +
  `${encounteredLegacyClasses.size} legacy classes remaining).`,
)
