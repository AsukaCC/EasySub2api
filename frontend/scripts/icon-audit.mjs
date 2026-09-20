import { readFileSync, readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import { parse } from 'vue/compiler-sfc'
import ts from 'typescript'

const root = fileURLToPath(new URL('../src', import.meta.url))
const iconSource = parse(readFileSync(path.join(root, 'components/icons/Icon.vue'), 'utf8')).descriptor.scriptSetup.content
const ast = ts.createSourceFile('Icon.ts', iconSource, ts.ScriptTarget.Latest, true)
const names = new Set()
function collect(node) {
  if (ts.isVariableDeclaration(node) && node.name.getText(ast) === 'icons') {
    let value = node.initializer
    while (ts.isAsExpression(value)) value = value.expression
    for (const property of value.properties) names.add(property.name.text)
  }
  ts.forEachChild(node, collect)
}
collect(ast)
const sizes = new Set(['xs', 'sm', 'md', 'lg', 'xl'])
const errors = []
let references = 0
let dynamicReferences = 0
function check(file) {
  const { descriptor } = parse(readFileSync(file, 'utf8'))
  function visit(node) {
    if (node.tag === 'Icon') {
      references++
      for (const prop of node.props || []) {
        if (prop.type === 6 && ['name', 'size'].includes(prop.name)) {
          const values = prop.name === 'name' ? names : sizes
          if (!values.has(prop.value?.content)) errors.push(`${path.relative(root, file)}:${node.loc.start.line} invalid ${prop.name}=${prop.value?.content}`)
        }
        if (prop.type === 7 && prop.name === 'bind' && prop.arg?.content === 'name') {
          dynamicReferences++
          const expression = ts.createSourceFile('expression.ts', prop.exp?.content || '', ts.ScriptTarget.Latest, true)
          const literals = (n) => {
            // Conditional icon expressions are checked here; variable/function types are checked by vue-tsc.
            if (ts.isConditionalExpression(n)) {
              for (const branch of [n.whenTrue, n.whenFalse]) {
                if (ts.isStringLiteral(branch) && !names.has(branch.text)) errors.push(`${path.relative(root, file)}:${node.loc.start.line} unknown dynamic icon ${branch.text}`)
              }
            }
            ts.forEachChild(n, literals)
          }
          literals(expression)
        }
      }
    }
    for (const child of node.children || []) visit(child)
  }
  if (descriptor.template?.ast) visit(descriptor.template.ast)
}
function walk(directory) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const file = path.join(directory, entry.name)
    if (entry.isDirectory()) walk(file)
    else if (entry.name.endsWith('.vue')) check(file)
  }
}
walk(root)
if (errors.length) { console.error(errors.join('\n')); process.exitCode = 1 }
else console.log(`Icon audit passed: ${names.size} icons, ${references} references (${dynamicReferences} dynamic).`)
