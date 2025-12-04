import path from 'node:path';
import { fileURLToPath } from 'node:url';

import globals from 'globals';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsParser from '@typescript-eslint/parser';
import importPlugin from 'eslint-plugin-import';
import unusedImports from 'eslint-plugin-unused-imports';
import promisePlugin from 'eslint-plugin-promise';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default [
  {
    ignores: ['dist/**', 'node_modules/**'],
  },
  {
    files: ['pkg/cluster/typescript/**/*.ts'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        project: './tsconfig.json',
        tsconfigRootDir: __dirname,
      },
      globals: {
        ...globals.node,
        ...globals.es2021,
      },
    },
    settings: {
      'import/resolver': {
        typescript: {
          project: './tsconfig.json',
        },
        node: {
          extensions: ['.js', '.ts', '.d.ts', '.json'],
        },
      },
    },
    plugins: {
      '@typescript-eslint': tseslint,
      import: importPlugin,
      'unused-imports': unusedImports,
      promise: promisePlugin,
    },
    rules: {
      // Base TS rules
      ...(tseslint.configs.recommended?.rules ?? {}),
      ...(tseslint.configs.stylistic?.rules ?? {}),

      // --- eslint-plugin-import ---
      'import/no-unresolved': 'error',
      'import/named': 'error',
      'import/no-duplicates': 'error',
      'import/order': [
        'warn',
        {
          groups: ['builtin', 'external', 'internal', 'parent', 'sibling', 'index', 'object', 'type'],
          'newlines-between': 'always',
          alphabetize: { order: 'asc', caseInsensitive: true },
        },
      ],

      // --- eslint-plugin-unused-imports ---
      'unused-imports/no-unused-imports': 'error',
      'unused-imports/no-unused-vars': [
        'warn',
        {
          vars: 'all',
          varsIgnorePattern: '^_',
          args: 'after-used',
          argsIgnorePattern: '^_',
        },
      ],

      // --- eslint-plugin-promise ---
      'promise/param-names': 'error',
      'promise/no-return-wrap': 'error',
      'promise/no-nesting': 'warn',
      'promise/no-new-statics': 'error',
      'promise/no-return-in-finally': 'warn',

      // Stylistic preference
      '@typescript-eslint/consistent-type-definitions': ['error', 'type'],
    },
  },
];
