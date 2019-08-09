const path = require('path')
// import Purgecss webpack plugin and glob-all
const PurgecssPlugin = require('purgecss-webpack-plugin')
const glob = require('glob-all')
const webpack = require('webpack')
const TerserPlugin = require('terser-webpack-plugin')

const cssWhiteList = []
const cssWhiteListPatterns = [/^simplebar/, /^cxlt-vue2-toastr/, /^ql-/, /^toast/]

module.exports = {
    configureWebpack: {
        plugins: [
            // Remove unused CSS using purgecss. See https://github.com/FullHuman/purgecss
            // for more information about purgecss.
            new PurgecssPlugin({
                paths: glob.sync([
                    path.join(__dirname, './../public/index.html'),
                    path.join(__dirname, './../**/*.vue'),
                    path.join(__dirname, './../src/**/*.js')
                ]),
                whitelist: cssWhiteList,
                whitelistPatterns: cssWhiteListPatterns
            }),
            new webpack.ProvidePlugin({
                'window.Quill': 'quill/dist/quill.js'
            })
        ],
        performance: {
            hints: false
        },
        optimization: {
            minimizer: [new TerserPlugin({
                cache: true,
                parallel: true,
                sourceMap: false,
                terserOptions: {
                    output: {
                        comments: false
                    }
                }
            })]
        }
    },

    assetsDir: 'static',
    productionSourceMap: false,
    outputDir: 'templates'

}
