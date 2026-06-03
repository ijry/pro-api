import{ap as N,x as f,y as L,al as U,aU as k,bp as A,c1 as E,b as H,cu as O,cI as S,a3 as y,ac as $,bV as z,aJ as l}from"./index-ClKeNBH9.js";import{u as j}from"./use-houdini-BzACH5SR.js";function I(e){const{heightSmall:a,heightMedium:n,heightLarge:i,borderRadius:s}=e;return{color:"#eee",colorEnd:"#ddd",borderRadius:s,heightSmall:a,heightMedium:n,heightLarge:i}}const M={common:N,self:I},T=f([L("skeleton",`
 height: 1em;
 width: 100%;
 transition:
 --n-color-start .3s var(--n-bezier),
 --n-color-end .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 animation: 2s skeleton-loading infinite cubic-bezier(0.36, 0, 0.64, 1);
 background-color: var(--n-color-start);
 `),f("@keyframes skeleton-loading",`
 0% {
 background: var(--n-color-start);
 }
 40% {
 background: var(--n-color-end);
 }
 80% {
 background: var(--n-color-start);
 }
 100% {
 background: var(--n-color-start);
 }
 `)]),V=Object.assign(Object.assign({},S.props),{text:Boolean,round:Boolean,circle:Boolean,height:[String,Number],width:[String,Number],size:String,repeat:{type:Number,default:1},animated:{type:Boolean,default:!0},sharp:{type:Boolean,default:!0}}),K=U({name:"Skeleton",inheritAttrs:!1,props:V,setup(e){j();const{mergedClsPrefixRef:a,mergedComponentPropsRef:n}=O(e),i=y(()=>{var t,o;return e.size||((o=(t=n==null?void 0:n.value)===null||t===void 0?void 0:t.Skeleton)===null||o===void 0?void 0:o.size)}),s=S("Skeleton","-skeleton",T,M,e,a);return{mergedClsPrefix:a,style:y(()=>{var t,o;const h=s.value,{common:{cubicBezierEaseInOut:x}}=h,b=h.self,{color:R,colorEnd:P,borderRadius:_}=b;let d;const{circle:c,sharp:w,round:B,width:r,height:u,text:p,animated:C}=e,v=i.value;v!==void 0&&(d=b[$("height",v)]);const m=c?(t=r??u)!==null&&t!==void 0?t:d:r,g=(o=c?r??u:u)!==null&&o!==void 0?o:d;return{display:p?"inline-block":"",verticalAlign:p?"-0.125em":"",borderRadius:c?"50%":B?"4096px":w?"":_,width:typeof m=="number"?z(m):m,height:typeof g=="number"?z(g):g,animation:C?"":"none","--n-bezier":x,"--n-color-start":R,"--n-color-end":P}})}},render(){const{repeat:e,style:a,mergedClsPrefix:n,$attrs:i}=this,s=k("div",A({class:`${n}-skeleton`,style:a},i));return e>1?k(H,null,E(e,null).map(t=>[s,`
`])):s}}),W={overview:()=>l("/api/admin/stats/overview"),timeseries:e=>l("/api/admin/stats/timeseries",e),byModel:e=>l("/api/admin/stats/by_model",e),byChannel:e=>l("/api/admin/stats/by_channel",e),byUser:e=>l("/api/admin/stats/by_user",e),exportURL:e=>`/api/admin/stats/export?${new URLSearchParams({format:"csv",...e}).toString()}`};export{K as N,W as s};
