import js from '@eslint/js'
import globals from 'globals'
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks'
import prettier from 'eslint-plugin-prettier';

export default [
  { ignores: ['dist'] },
  {
    files: ['**/*.{js,jsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: {
        ...globals.browser,
        varnishmon: 'readonly',
      },
      parserOptions: {
        ecmaVersion: 'latest',
        ecmaFeatures: { jsx: true },
        sourceType: 'module',
      },
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
    plugins: {
      'react': react,
      'react-hooks': reactHooks,
      'prettier': prettier,
    },
    rules: {
      ...js.configs.recommended.rules,
      ...react.configs.recommended.rules,
      ...reactHooks.configs.recommended.rules,
      ...prettier.configs.recommended.rules,
      'no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_' },
      ],
      eqeqeq: [
        'error',
        'smart',
      ],
    },
  },
]
