const isProd = process.env.NODE_ENV === "production";

const withNextra = require('nextra')({
    theme: 'nextra-theme-docs',
    themeConfig: './theme.config.js',
    unstable_staticImage: true,
})
module.exports = withNextra({ 
    basePath: "",
    assetPrefix: "", experimental: {
    images: {
        unoptimized: true
    }
}})