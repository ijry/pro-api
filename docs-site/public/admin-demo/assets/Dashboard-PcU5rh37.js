import{x as I,y as k,D as B,A as w,a$ as J,b0 as W,al as Q,aU as q,cu as X,cF as Y,cI as G,cJ as ee,bm as te,a3 as re,ab as oe,bU as se,cn as le,aZ as ae,cj as ie,bI as ne,a9 as x,a6 as r,ag as i,cP as l,cs as e,bZ as R,bK as s,f as C,a7 as u,cm as c,by as M,b as T,b$ as F,cE as de,ae as ue}from"./index-ClKeNBH9.js";import{s as E,N as z}from"./stats-BDj45DcB.js";import{n as ce}from"./notice-U7Pvbskc.js";import{_ as ve}from"./TimeDisplay.vue_vue_type_script_setup_true_lang-Cxtwypwx.js";import{N as K,a as $}from"./Grid-DnL_Wmf5.js";import{N as O}from"./Statistic-BhY90LS6.js";import{N as V}from"./Empty-DQjXSwmQ.js";import{N as pe}from"./text-CiDnAQYY.js";import"./use-houdini-BzACH5SR.js";import"./dayjs.min-DSLg8duS.js";import"./Tooltip-jxqJeGHW.js";import"./Popover-Dm198j0w.js";import"./get-48VdzrSm.js";import"./cssr-DxXR4Bge.js";import"./get-slot-Bk_rJcZu.js";import"./use-locale-CGm--TNI.js";const me=I([k("list",`
 --n-merged-border-color: var(--n-border-color);
 --n-merged-color: var(--n-color);
 --n-merged-color-hover: var(--n-color-hover);
 margin: 0;
 font-size: var(--n-font-size);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 padding: 0;
 list-style-type: none;
 color: var(--n-text-color);
 background-color: var(--n-merged-color);
 `,[B("show-divider",[k("list-item",[I("&:not(:last-child)",[w("divider",`
 background-color: var(--n-merged-border-color);
 `)])])]),B("clickable",[k("list-item",`
 cursor: pointer;
 `)]),B("bordered",`
 border: 1px solid var(--n-merged-border-color);
 border-radius: var(--n-border-radius);
 `),B("hoverable",[k("list-item",`
 border-radius: var(--n-border-radius);
 `,[I("&:hover",`
 background-color: var(--n-merged-color-hover);
 `,[w("divider",`
 background-color: transparent;
 `)])])]),B("bordered, hoverable",[k("list-item",`
 padding: 12px 20px;
 `),w("header, footer",`
 padding: 12px 20px;
 `)]),w("header, footer",`
 padding: 12px 0;
 box-sizing: border-box;
 transition: border-color .3s var(--n-bezier);
 `,[I("&:not(:last-child)",`
 border-bottom: 1px solid var(--n-merged-border-color);
 `)]),k("list-item",`
 position: relative;
 padding: 12px 0; 
 box-sizing: border-box;
 display: flex;
 flex-wrap: nowrap;
 align-items: center;
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `,[w("prefix",`
 margin-right: 20px;
 flex: 0;
 `),w("suffix",`
 margin-left: 20px;
 flex: 0;
 `),w("main",`
 flex: 1;
 `),w("divider",`
 height: 1px;
 position: absolute;
 bottom: 0;
 left: 0;
 right: 0;
 background-color: transparent;
 transition: background-color .3s var(--n-bezier);
 pointer-events: none;
 `)])]),J(k("list",`
 --n-merged-color-hover: var(--n-color-hover-modal);
 --n-merged-color: var(--n-color-modal);
 --n-merged-border-color: var(--n-border-color-modal);
 `)),W(k("list",`
 --n-merged-color-hover: var(--n-color-hover-popover);
 --n-merged-color: var(--n-color-popover);
 --n-merged-border-color: var(--n-border-color-popover);
 `))]),be=Object.assign(Object.assign({},G.props),{size:{type:String,default:"medium"},bordered:Boolean,clickable:Boolean,hoverable:Boolean,showDivider:{type:Boolean,default:!0}}),Z=oe("n-list"),fe=Q({name:"List",props:be,slots:Object,setup(n){const{mergedClsPrefixRef:a,inlineThemeDisabled:o,mergedRtlRef:N}=X(n),D=Y("List",N,a),j=G("List","-list",me,te,n,a);se(Z,{showDividerRef:le(n,"showDivider"),mergedClsPrefixRef:a});const L=re(()=>{const{common:{cubicBezierEaseInOut:v},self:{fontSize:H,textColor:U,color:A,colorModal:P,colorPopover:S,borderColor:m,borderColorModal:d,borderColorPopover:t,borderRadius:p,colorHover:b,colorHoverModal:f,colorHoverPopover:h}}=j.value;return{"--n-font-size":H,"--n-bezier":v,"--n-text-color":U,"--n-color":A,"--n-border-radius":p,"--n-border-color":m,"--n-border-color-modal":d,"--n-border-color-popover":t,"--n-color-modal":P,"--n-color-popover":S,"--n-color-hover":b,"--n-color-hover-modal":f,"--n-color-hover-popover":h}}),_=o?ee("list",void 0,L,n):void 0;return{mergedClsPrefix:a,rtlEnabled:D,cssVars:o?void 0:L,themeClass:_==null?void 0:_.themeClass,onRender:_==null?void 0:_.onRender}},render(){var n;const{$slots:a,mergedClsPrefix:o,onRender:N}=this;return N==null||N(),q("ul",{class:[`${o}-list`,this.rtlEnabled&&`${o}-list--rtl`,this.bordered&&`${o}-list--bordered`,this.showDivider&&`${o}-list--show-divider`,this.hoverable&&`${o}-list--hoverable`,this.clickable&&`${o}-list--clickable`,this.themeClass],style:this.cssVars},a.header?q("div",{class:`${o}-list__header`},a.header()):null,(n=a.default)===null||n===void 0?void 0:n.call(a),a.footer?q("div",{class:`${o}-list__footer`},a.footer()):null)}}),he=Q({name:"ListItem",slots:Object,setup(){const n=ae(Z,null);return n||ie("list-item","`n-list-item` must be placed in `n-list`."),{showDivider:n.showDividerRef,mergedClsPrefix:n.mergedClsPrefixRef}},render(){const{$slots:n,mergedClsPrefix:a}=this;return q("li",{class:`${a}-list-item`},n.prefix?q("div",{class:`${a}-list-item__prefix`},n.prefix()):null,n.default?q("div",{class:`${a}-list-item__main`},n):null,n.suffix?q("div",{class:`${a}-list-item__suffix`},n.suffix()):null,this.showDivider&&q("div",{class:`${a}-list-item__divider`}))}}),ge={key:2,class:"w-full text-sm"},ye={class:"py-1 font-mono text-xs"},xe={class:"py-1 text-right"},_e={class:"py-1 text-right"},ke={key:2,class:"w-full text-sm"},we={class:"py-1"},Ce={class:"py-1 text-right"},ze={class:"py-1 text-right"},$e={key:2,class:"w-full text-sm"},qe={class:"py-1"},Ne={class:"py-1 text-right"},Re={class:"py-1 text-right"},Le={class:"flex items-center justify-between"},Qe=Q({__name:"Dashboard",setup(n){const a=de(),o=R(null),N=R([]),D=R([]),j=R([]),L=R([]),_=R([]),v=R({overview:!0,charts:!0,notices:!0});async function H(){v.value.overview=!0;try{o.value=await E.overview()}catch{}finally{v.value.overview=!1}}async function U(){v.value.charts=!0;try{const[m,d,t,p]=await Promise.all([E.timeseries({granularity:"hour"}),E.byModel({order_by:"quota",limit:10}),E.byChannel({order_by:"requests",limit:10}),E.byUser({order_by:"quota",limit:10})]);N.value=m.points,D.value=d.rows,j.value=t.rows,L.value=p.rows}catch{}finally{v.value.charts=!1}}async function A(){v.value.notices=!0;try{const m=await ce.list({status:1,page:1,size:5});_.value=m.items}catch{}finally{v.value.notices=!1}}ne(()=>{H(),U(),A()});const P=m=>m>=0?"text-green-500":"text-red-500",S=m=>m>=0?"+":"";return(m,d)=>(s(),x("div",null,[d[7]||(d[7]=r("h2",{class:"text-2xl font-semibold mb-4"},"仪表盘",-1)),i(e(K),{cols:4,"x-gap":16,"y-gap":16,responsive:"screen","item-responsive":!0,class:"mb-4"},{default:l(()=>[i(e($),{span:"4 600:2 900:1"},{default:l(()=>[i(e(C),{size:"small",hoverable:"",onClick:d[0]||(d[0]=t=>e(a).push("/logs/requests"))},{default:l(()=>{var t;return[v.value.overview?(s(),u(e(z),{key:0,text:"",repeat:2})):(s(),u(e(O),{key:1,label:"今日请求",value:((t=o.value)==null?void 0:t.requests_today)??0},{suffix:l(()=>{var p,b,f,h,g,y;return[r("span",{class:M(P(((b=(p=o.value)==null?void 0:p.delta)==null?void 0:b.requests)??0))},c(S(((h=(f=o.value)==null?void 0:f.delta)==null?void 0:h.requests)??0))+c(((((y=(g=o.value)==null?void 0:g.delta)==null?void 0:y.requests)??0)*100).toFixed(1))+"% ",3)]}),_:1},8,["value"]))]}),_:1})]),_:1}),i(e($),{span:"4 600:2 900:1"},{default:l(()=>[i(e(C),{size:"small",hoverable:"",onClick:d[1]||(d[1]=t=>e(a).push("/stats"))},{default:l(()=>{var t;return[v.value.overview?(s(),u(e(z),{key:0,text:"",repeat:2})):(s(),u(e(O),{key:1,label:"今日收入(quota)",value:(((t=o.value)==null?void 0:t.revenue_today)??0).toLocaleString()},{suffix:l(()=>{var p,b,f,h,g,y;return[r("span",{class:M(P(((b=(p=o.value)==null?void 0:p.delta)==null?void 0:b.revenue)??0))},c(S(((h=(f=o.value)==null?void 0:f.delta)==null?void 0:h.revenue)??0))+c(((((y=(g=o.value)==null?void 0:g.delta)==null?void 0:y.revenue)??0)*100).toFixed(1))+"% ",3)]}),_:1},8,["value"]))]}),_:1})]),_:1}),i(e($),{span:"4 600:2 900:1"},{default:l(()=>[i(e(C),{size:"small",hoverable:"",onClick:d[2]||(d[2]=t=>e(a).push("/users"))},{default:l(()=>{var t;return[v.value.overview?(s(),u(e(z),{key:0,text:"",repeat:2})):(s(),u(e(O),{key:1,label:"活跃用户",value:((t=o.value)==null?void 0:t.active_users)??0},{suffix:l(()=>{var p,b,f,h,g,y;return[r("span",{class:M(P(((b=(p=o.value)==null?void 0:p.delta)==null?void 0:b.users)??0))},c(S(((h=(f=o.value)==null?void 0:f.delta)==null?void 0:h.users)??0))+c(((((y=(g=o.value)==null?void 0:g.delta)==null?void 0:y.users)??0)*100).toFixed(1))+"% ",3)]}),_:1},8,["value"]))]}),_:1})]),_:1}),i(e($),{span:"4 600:2 900:1"},{default:l(()=>[i(e(C),{size:"small",hoverable:"",onClick:d[3]||(d[3]=t=>e(a).push("/logs/errors"))},{default:l(()=>{var t;return[v.value.overview?(s(),u(e(z),{key:0,text:"",repeat:2})):(s(),u(e(O),{key:1,label:"错误率",value:`${((((t=o.value)==null?void 0:t.error_rate)??0)*100).toFixed(2)}%`},{suffix:l(()=>{var p,b,f,h,g,y;return[r("span",{class:M(P(-(((b=(p=o.value)==null?void 0:p.delta)==null?void 0:b.error_rate)??0)))},c(S(-(((h=(f=o.value)==null?void 0:f.delta)==null?void 0:h.error_rate)??0)))+c(((((y=(g=o.value)==null?void 0:g.delta)==null?void 0:y.error_rate)??0)*100).toFixed(2))+"% ",3)]}),_:1},8,["value"]))]}),_:1})]),_:1})]),_:1}),i(e(K),{cols:2,"x-gap":16,"y-gap":16,responsive:"screen","item-responsive":!0,class:"mb-4"},{default:l(()=>[i(e($),{span:"2 900:1"},{default:l(()=>[i(e(C),{title:"Top 模型 (by quota)",size:"small"},{default:l(()=>[v.value.charts?(s(),u(e(z),{key:0,text:"",repeat:5})):D.value.length?(s(),x("table",ge,[d[4]||(d[4]=r("thead",null,[r("tr",null,[r("th",{class:"text-left py-1"},"模型"),r("th",{class:"text-right py-1"},"请求数"),r("th",{class:"text-right py-1"},"Quota")])],-1)),r("tbody",null,[(s(!0),x(T,null,F(D.value,t=>(s(),x("tr",{key:t.model,class:"border-t border-gray-100 dark:border-gray-800"},[r("td",ye,c(t.model),1),r("td",xe,c(t.requests.toLocaleString()),1),r("td",_e,c(t.quota.toLocaleString()),1)]))),128))])])):(s(),u(e(V),{key:1,description:"暂无数据"}))]),_:1})]),_:1}),i(e($),{span:"2 900:1"},{default:l(()=>[i(e(C),{title:"Top 渠道 (by requests)",size:"small"},{default:l(()=>[v.value.charts?(s(),u(e(z),{key:0,text:"",repeat:5})):j.value.length?(s(),x("table",ke,[d[5]||(d[5]=r("thead",null,[r("tr",null,[r("th",{class:"text-left py-1"},"渠道"),r("th",{class:"text-right py-1"},"请求数"),r("th",{class:"text-right py-1"},"Quota")])],-1)),r("tbody",null,[(s(!0),x(T,null,F(j.value,t=>(s(),x("tr",{key:t.channel_id,class:"border-t border-gray-100 dark:border-gray-800"},[r("td",we,c(t.channel_name),1),r("td",Ce,c(t.requests.toLocaleString()),1),r("td",ze,c(t.quota.toLocaleString()),1)]))),128))])])):(s(),u(e(V),{key:1,description:"暂无数据"}))]),_:1})]),_:1})]),_:1}),i(e(K),{cols:2,"x-gap":16,"y-gap":16,responsive:"screen","item-responsive":!0},{default:l(()=>[i(e($),{span:"2 900:1"},{default:l(()=>[i(e(C),{title:"Top 用户 (by quota)",size:"small"},{default:l(()=>[v.value.charts?(s(),u(e(z),{key:0,text:"",repeat:5})):L.value.length?(s(),x("table",$e,[d[6]||(d[6]=r("thead",null,[r("tr",null,[r("th",{class:"text-left py-1"},"用户"),r("th",{class:"text-right py-1"},"请求数"),r("th",{class:"text-right py-1"},"Quota")])],-1)),r("tbody",null,[(s(!0),x(T,null,F(L.value,t=>(s(),x("tr",{key:t.user_id,class:"border-t border-gray-100 dark:border-gray-800"},[r("td",qe,c(t.username),1),r("td",Ne,c(t.requests.toLocaleString()),1),r("td",Re,c(t.quota.toLocaleString()),1)]))),128))])])):(s(),u(e(V),{key:1,description:"暂无数据"}))]),_:1})]),_:1}),i(e($),{span:"2 900:1"},{default:l(()=>[i(e(C),{title:"最新公告",size:"small"},{default:l(()=>[v.value.notices?(s(),u(e(z),{key:0,text:"",repeat:5})):_.value.length?(s(),u(e(fe),{key:2},{default:l(()=>[(s(!0),x(T,null,F(_.value,t=>(s(),u(e(he),{key:t.id},{default:l(()=>[r("div",Le,[i(e(pe),null,{default:l(()=>[ue(c(t.title),1)]),_:2},1024),i(ve,{value:t.publish_at,relative:"",class:"text-xs opacity-50"},null,8,["value"])])]),_:2},1024))),128))]),_:1})):(s(),u(e(V),{key:1,description:"暂无公告"}))]),_:1})]),_:1})]),_:1})]))}});export{Qe as default};
