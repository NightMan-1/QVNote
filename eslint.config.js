import vue from 'eslint-plugin-vue'

export default [
    {
        ignores: ['dist/**', 'templates/**', 'public/static/**', 'node_modules/**']
    },
    ...vue.configs['flat/essential'],
    {
        files: ['**/*.{js,vue}'],
        languageOptions: {
            ecmaVersion: 2022,
            sourceType: 'module'
        },
        rules: {
            indent: ['error', 4],
            'no-tabs': 0,
            'no-mixed-spaces-and-tabs': 0,
            'new-cap': 0,
            'generator-star-spacing': 'off',
            'vue/multi-word-component-names': 'off'
        }
    }
]
