(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[405],{1464:function(e,n,t){(window.__NEXT_P=window.__NEXT_P||[]).push(["/",function(){return t(9124)}])},7845:function(e,n,t){"use strict";var r=t(5893);n.Z={github:"https://github.com/ionos-cloud/octopinger",docsRepositoryBase:"https://github.com/ionos-cloud/octopinger/blob/main/docs/pages",titleSuffix:" \u2013 Run",nextLinks:!0,prevLinks:!0,search:!0,customSearch:null,darkMode:!0,footer:!0,footerText:"Apache-2.0 ".concat((new Date).getFullYear()," \xa9 IONOS SE."),footerEditLink:"Edit this page on GitHub",logo:(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)("svg",{children:"..."}),(0,r.jsx)("span",{children:"\ud83d\udc19 Octopinger"})]})}},9124:function(e,n,t){"use strict";t.r(n);var r=t(5893),o=t(7829),i=t.n(o),a=t(3805),s=t(7845),l=(t(5675),t(1132),t(1127)),c=t.n(l),d=function(e){return(0,a.withSSG)(i()({filename:"index.mdx",route:"/",meta:{title:"Run - A versatile task runner"},pageMap:[{name:"index",route:"/",frontMatter:{title:"Run - A versatile task runner"}},{name:"meta.json",meta:{index:{title:"Introduction",type:"page",hidden:!0}}},{name:"metrics",route:"/metrics",frontMatter:{title:"Metrics"}}]},s.Z))(e)};function p(e){var n=Object.assign({h1:"h1",p:"p",a:"a",pre:"pre",code:"code",h2:"h2",ul:"ul",li:"li"},e.components);return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(n.h1,{children:"\ud83d\udc19 Octopinger"}),"\n",(0,r.jsxs)(n.p,{children:["Octopinger is an Kubernetes Operator to monitor the connectivity of your cluster. The probes use ICMP to determine the connectivity between cluster nodes. Metrics are exported via ",(0,r.jsx)(n.a,{href:"https://prometheus.io/",children:"Prometheus"}),"."]}),"\n",(0,r.jsx)(c(),{children:(0,r.jsx)(n.p,{children:"This is under active development."})}),"\n",(0,r.jsxs)(n.p,{children:["Installation is super easy. You can use ",(0,r.jsx)(n.a,{href:"https://helm.sh/",children:"Helm"})," to install the operator to you cluster and create Octopinger instances."]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-bash",children:"helm repo add octopinger https://octopinger.io/\nhelm repo update \n"})}),"\n",(0,r.jsxs)(n.p,{children:["Install Octopinger to your cluster in a ",(0,r.jsx)(n.code,{children:"octopinger"})," namespace."]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-bash",children:"helm install octopinger octopinger/octopinger --create-namespace --namespace octopinger\n"})}),"\n",(0,r.jsx)(n.p,{children:"After the installation you can use this example to create an Octopinger."}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-yaml",children:"apiVersion: octopinger.io/v1alpha1\nkind: Octopinger\nmetadata:\n  name: demo\nspec:\n  label: octopinger\n  config:\n    icmp:\n      enable: true\n    dns:\n      enable: false\n  template:\n    image: ghcr.io/ionos-cloud/octopinger/octopinger:v0.1.6\n"})}),"\n",(0,r.jsx)(n.h2,{children:"Features"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsx)(n.li,{children:"Ping all nodes in your cluster to detect availability and network issues"}),"\n",(0,r.jsx)(n.li,{children:"Test domain name record availability"}),"\n",(0,r.jsxs)(n.li,{children:["Many ",(0,r.jsx)(n.a,{href:"/metrics",children:"metrics"})," exposed via ",(0,r.jsx)(n.a,{href:"https://prometheus.io/",children:"Prometheus"})]}),"\n"]})]})}n.default=function(){var e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:{};return(0,r.jsx)(d,Object.assign({},e,{children:(0,r.jsx)(p,e)}))}},1132:function(e,n,t){t(3045)},1127:function(e,n,t){e.exports=t(3952)},3045:function(e,n,t){function r(e){return e&&"object"===typeof e&&"default"in e?e.default:e}var o=r(t(7294)),i=r(t(4184));e.exports=({full:e,children:n})=>o.createElement("div",{className:i("bleed relative mt-6 -mx-6 md:-mx-8 2xl:-mx-24",{full:e})},n)},3952:function(e,n,t){var r,o=(r=t(7294))&&"object"===typeof r&&"default"in r?r.default:r;const i={default:"bg-orange-100 text-orange-800 dark:text-orange-300 dark:bg-orange-200 dark:bg-opacity-10",error:"bg-red-200 text-red-900 dark:text-red-200 dark:bg-red-600 dark:bg-opacity-30",warning:"bg-yellow-200 text-yellow-900 dark:text-yellow-200 dark:bg-yellow-700 dark:bg-opacity-30"};e.exports=({children:e,type:n="default",emoji:t="\ud83d\udca1"})=>o.createElement("div",{className:`${i[n]} flex rounded-lg callout mt-6`},o.createElement("div",{className:"pl-3 pr-2 py-2 select-none text-xl",style:{fontFamily:'"Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"'}},t),o.createElement("div",{className:"pr-4 py-2"},e))}},function(e){e.O(0,[58,774,888,179],(function(){return n=1464,e(e.s=n);var n}));var n=e.O();_N_E=n}]);