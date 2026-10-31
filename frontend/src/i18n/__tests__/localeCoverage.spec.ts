import { readFileSync, readdirSync } from 'node:fs'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'
import ts from 'typescript'
import { compileTemplate, parse } from 'vue/compiler-sfc'
import { describe, expect, it } from 'vitest'
import en from '../locales/en'
import zh from '../locales/zh'

function messageKeys(messages: object, prefix = ''): string[] {
  return Object.entries(messages).flatMap(([key, value]) => {
    const path = prefix ? `${prefix}.${key}` : key
    return typeof value === 'string' ? [path] : messageKeys(value, path)
  })
}

const enKeys = new Set(messageKeys(en))
const zhKeys = new Set(messageKeys(zh))
const sourceRoot = join(dirname(fileURLToPath(import.meta.url)), '../..')

function sourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) {
      return ['__tests__', 'i18n'].includes(entry.name) ? [] : sourceFiles(path)
    }
    return /\.(vue|ts)$/.test(entry.name) && !/\.(spec|test|d)\.ts$/.test(entry.name) ? [path] : []
  })
}

function translationReferences(source: string): string[] {
  const keys: string[] = []
  const tree = ts.createSourceFile('source.ts', source, ts.ScriptTarget.Latest, true)
  function visit(node: ts.Node) {
    if (ts.isCallExpression(node)) {
      const callee = node.expression
      const name = ts.isIdentifier(callee) ? callee.text : ts.isPropertyAccessExpression(callee) ? callee.name.text : ''
      const key = node.arguments[0]
      if (['t', '$t'].includes(name) && key && (ts.isStringLiteral(key) || ts.isNoSubstitutionTemplateLiteral(key))) {
        keys.push(key.text)
      }
    }
    ts.forEachChild(node, visit)
  }
  visit(tree)
  return keys
}

describe('locale coverage', () => {
  it('provides the same message keys in both languages', () => {
    expect([...enKeys].sort()).toEqual([...zhKeys].sort())
  })

  it('resolves static translation references in application scripts and templates', () => {
    const missing: string[] = []
    for (const filename of sourceFiles(sourceRoot)) {
      const content = readFileSync(filename, 'utf8')
      let sources = [content]
      if (filename.endsWith('.vue')) {
        const { descriptor } = parse(content, { filename })
        sources = [descriptor.script?.content ?? '', descriptor.scriptSetup?.content ?? '']
        if (descriptor.template) {
          // Compile expressions so directives and interpolations are checked without scanning comments or text.
          sources.push(compileTemplate({ source: descriptor.template.content, filename, id: filename }).code)
        }
      }
      for (const key of new Set(sources.flatMap(translationReferences))) {
        if (!enKeys.has(key) || !zhKeys.has(key)) missing.push(`${relative(sourceRoot, filename)}: ${key}`)
      }
    }
    expect(missing).toEqual([])
  })
})
