const path = require('path')
// import Purgecss webpack plugin and glob-all
const PurgecssPlugin = require('purgecss-webpack-plugin')
const glob = require('glob-all')
const webpack = require('webpack')
const TerserPlugin = require('terser-webpack-plugin')
// const { DuplicatesPlugin } = require("inspectpack/plugin")

const cssWhiteList = []
const cssWhiteListPatterns = []
const whitelistPatternsChildren = [/vch/, /^ql-/, /^toast/, /^tingle/]

module.exports = {
    configureWebpack: {
        plugins: [
            /*
            new DuplicatesPlugin({
                // Emit compilation warning or error? (Default: `false`)
                emitErrors: false,
                // Handle all messages with handler function (`(report: string)`)
                // Overrides `emitErrors` output.
                emitHandler: undefined,
                // Display full duplicates information? (Default: `false`)
                verbose: false
            }), */
            // Remove unused CSS using purgecss. See https://github.com/FullHuman/purgecss
            // for more information about purgecss.
            new PurgecssPlugin({
                paths: glob.sync([
                    path.join(__dirname, './../public/index.html'),
                    path.join(__dirname, './../**/*.vue'),
                    path.join(__dirname, './../src/**/*.js')
                ]),
                whitelist: cssWhiteList,
                whitelistPatterns: cssWhiteListPatterns,
                whitelistPatternsChildren: whitelistPatternsChildren
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
