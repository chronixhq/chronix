import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import unusedImports from 'eslint-plugin-unused-imports'
import tseslint from 'typescript-eslint'

export default tseslint.config(
	{ignores: ['dist']},
	{
		files: ['**/*.{ts,tsx}'],
		extends: [js.configs.recommended, ...tseslint.configs.recommended],
		languageOptions: {
			ecmaVersion: 2020,
			globals: globals.browser,
		},
		plugins: {
			'react-refresh': reactRefresh,
			'react-hooks': reactHooks,
			'unused-imports': unusedImports,
		},
		rules: {
			// Enforce explicit type-only imports to prevent runtime import errors and improve tree-shaking
			'@typescript-eslint/consistent-type-imports': [
				'error', {
					prefer: 'type-imports',
					fixStyle: 'separate-type-imports',
					disallowTypeAnnotations: false
				}
			],
			'@typescript-eslint/no-explicit-any': 'off',
			'unused-imports/no-unused-imports': 'error',
			'react-refresh/only-export-components': ['warn', {allowConstantExport: true}],
			'react-hooks/rules-of-hooks': 'error',
			'react-hooks/exhaustive-deps': 'warn',
			'no-empty': ['error', {allowEmptyCatch: true}],
		},
	},
	{
		files: ['src/**/*.{tsx,jsx}'],
		rules: {
			'react-refresh/only-export-components': ['warn', {allowConstantExport: true}],
		},
	},
	{
		files: ['src/data/**/*.{ts,tsx}', 'src/lib/**/*.{ts,tsx}'],
		rules: {
			'react-refresh/only-export-components': 'off',
		},
	},
)
