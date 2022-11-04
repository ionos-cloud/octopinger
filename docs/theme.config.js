// theme.config.js
export default {
    github: 'https://github.com/ionos-cloud/octopinger',
    docsRepositoryBase: 'https://github.com/ionos-cloud/octopinger/blob/main/docs/pages', // base URL for the docs repository
    titleSuffix: ' – Run',
    nextLinks: true,
    prevLinks: true,
    search: true,
    customSearch: null, // customizable, you can use algolia for example
    darkMode: true,
    footer: true,
    footerText: `Apache-2.0 ${new Date().getFullYear()} © IONOS SE.`,
    footerEditLink: `Edit this page on GitHub`,
    logo: (
      <>
        <svg>...</svg>
        <span>🐙 Octopinger</span>
      </>
    )
  }