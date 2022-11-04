// theme.config.js
export default {
    github: 'https://github.com/katallaxie/run',
    docsRepositoryBase: 'https://github.com/katallaxie/run/blob/main/docs/pages', // base URL for the docs repository
    titleSuffix: ' – Run',
    nextLinks: true,
    prevLinks: true,
    search: true,
    customSearch: null, // customizable, you can use algolia for example
    darkMode: true,
    footer: true,
    footerText: `MIT ${new Date().getFullYear()} © Sebastian Doell (@katallaxie).`,
    footerEditLink: `Edit this page on GitHub`,
    logo: (
      <>
        <svg>...</svg>
        <span>Run</span>
      </>
    )
  }