import{aX as fe,a2 as C,a3 as ke,an as ge,x as R,y as v,D as P,A as B,ak as q,aS as s,cs as Y,bG as Q,cL as K,cG as V,cH as Z,bX as T,cl as ne,cc as Se,d as me,I as pe,W as he,E as ve,k as be,bu as Ne,bR as je,ab as L,aa as Pe,bS as Be,c6 as ie,c4 as oe,ch as Re,a6 as M,cq as n,cN as g,ad as x,ck as _,a5 as E,by as le,af as m,a8 as D,a7 as F,bI as N,cu as ee,b as ye,bZ as Ie,cz as Te,B as H,i as qe,f as U,cC as De,cB as We}from"./index-BMyk45kF.js";import{a as te,u as Oe}from"./useAccountActions-Ds5uy-2j.js";import{f as O}from"./get-D86kArqK.js";import{N as Le}from"./Tooltip-CP31tRHA.js";import{N as Me}from"./Pagination-D79ZlzbF.js";import{u as He}from"./use-houdini-DjaU2PqG.js";import{N as _e}from"./Spin-CA7RLRm7.js";import{N as Ee}from"./Empty-cLnj6R0M.js";import{N as J}from"./Tag-1HdK0peN.js";import{N as X}from"./Space-D5OWEM5p.js";import{N as Ve}from"./Alert-DZw_pwSa.js";import{a as Ge}from"./Input-C2HQZJje.js";import{N as Ae}from"./PageHeader-3Er8wAZB.js";import{a as ae,N as Fe}from"./Grid-CJ9vz7RE.js";import{N as se,a as k}from"./DescriptionsItem-CHUNg359.js";import"./Popover-BQ0-iUw6.js";import"./cssr-E7GQDMMK.js";import"./Select-CNTyub3p.js";import"./create-BM_mgYTJ.js";import"./FocusDetector-B9d0hO94.js";import"./happens-in-CM8LO42l.js";import"./use-locale-DWTMmuPp.js";import"./create-ref-setter-C4J8sofl.js";import"./get-slot-Bk_rJcZu.js";function Xe(e,r){const a=fe(ke,null);return C(()=>e.hljs||(a==null?void 0:a.mergedHljsRef.value))}function Ye(e){const{textColor2:r,fontSize:a,fontWeightStrong:o,textColor3:c}=e;return{textColor:r,fontSize:a,fontWeightStrong:o,"mono-3":"#a0a1a7","hue-1":"#0184bb","hue-2":"#4078f2","hue-3":"#a626a4","hue-4":"#50a14f","hue-5":"#e45649","hue-5-2":"#c91243","hue-6":"#986801","hue-6-2":"#c18401",lineNumberTextColor:c}}const Ke={common:ge,self:Ye},Ue=R([v("code",`
 font-size: var(--n-font-size);
 font-family: var(--n-font-family);
 `,[P("show-line-numbers",`
 display: flex;
 `),B("line-numbers",`
 user-select: none;
 padding-right: 12px;
 text-align: right;
 transition: color .3s var(--n-bezier);
 color: var(--n-line-number-text-color);
 `),P("word-wrap",[R("pre",`
 white-space: pre-wrap;
 word-break: break-all;
 `)]),R("pre",`
 margin: 0;
 line-height: inherit;
 font-size: inherit;
 font-family: inherit;
 `),R("[class^=hljs]",`
 color: var(--n-text-color);
 transition: 
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `)]),({props:e})=>{const r=`${e.bPrefix}code`;return[`${r} .hljs-comment,
 ${r} .hljs-quote {
 color: var(--n-mono-3);
 font-style: italic;
 }`,`${r} .hljs-doctag,
 ${r} .hljs-keyword,
 ${r} .hljs-formula {
 color: var(--n-hue-3);
 }`,`${r} .hljs-section,
 ${r} .hljs-name,
 ${r} .hljs-selector-tag,
 ${r} .hljs-deletion,
 ${r} .hljs-subst {
 color: var(--n-hue-5);
 }`,`${r} .hljs-literal {
 color: var(--n-hue-1);
 }`,`${r} .hljs-string,
 ${r} .hljs-regexp,
 ${r} .hljs-addition,
 ${r} .hljs-attribute,
 ${r} .hljs-meta-string {
 color: var(--n-hue-4);
 }`,`${r} .hljs-built_in,
 ${r} .hljs-class .hljs-title {
 color: var(--n-hue-6-2);
 }`,`${r} .hljs-attr,
 ${r} .hljs-variable,
 ${r} .hljs-template-variable,
 ${r} .hljs-type,
 ${r} .hljs-selector-class,
 ${r} .hljs-selector-attr,
 ${r} .hljs-selector-pseudo,
 ${r} .hljs-number {
 color: var(--n-hue-6);
 }`,`${r} .hljs-symbol,
 ${r} .hljs-bullet,
 ${r} .hljs-link,
 ${r} .hljs-meta,
 ${r} .hljs-selector-id,
 ${r} .hljs-title {
 color: var(--n-hue-2);
 }`,`${r} .hljs-emphasis {
 font-style: italic;
 }`,`${r} .hljs-strong {
 font-weight: var(--n-font-weight-strong);
 }`,`${r} .hljs-link {
 text-decoration: underline;
 }`]}]),Je=Object.assign(Object.assign({},V.props),{language:String,code:{type:String,default:""},trim:{type:Boolean,default:!0},hljs:Object,uri:Boolean,inline:Boolean,wordWrap:Boolean,showLineNumbers:Boolean,internalFontSize:Number,internalNoHighlight:Boolean}),Qe=q({name:"Code",props:Je,setup(e,{slots:r}){const{internalNoHighlight:a}=e,{mergedClsPrefixRef:o,inlineThemeDisabled:c}=Y(),t=T(null),l=a?{value:void 0}:Xe(e),d=(i,f,u)=>{const{value:z}=l;return!z||!(i&&z.getLanguage(i))?null:z.highlight(u?f.trim():f,{language:i}).value},h=C(()=>e.inline||e.wordWrap?!1:e.showLineNumbers),b=()=>{if(r.default)return;const{value:i}=t;if(!i)return;const{language:f}=e,u=e.uri?window.decodeURIComponent(e.code):e.code;if(f){const w=d(f,u,e.trim);if(w!==null){if(e.inline)i.innerHTML=w;else{const S=i.querySelector(".__code__");S&&i.removeChild(S);const j=document.createElement("pre");j.className="__code__",j.innerHTML=w,i.appendChild(j)}return}}if(e.inline){i.textContent=u;return}const z=i.querySelector(".__code__");if(z)z.textContent=u;else{const w=document.createElement("pre");w.className="__code__",w.textContent=u,i.innerHTML="",i.appendChild(w)}};Q(b),K(ne(e,"language"),b),K(ne(e,"code"),b),a||K(l,b);const y=V("Code","-code",Ue,Ke,e,o),$=C(()=>{const{common:{cubicBezierEaseInOut:i,fontFamilyMono:f},self:{textColor:u,fontSize:z,fontWeightStrong:w,lineNumberTextColor:S,"mono-3":j,"hue-1":I,"hue-2":G,"hue-3":W,"hue-4":A,"hue-5":$e,"hue-5-2":Ce,"hue-6":ze,"hue-6-2":we}}=y.value,{internalFontSize:re}=e;return{"--n-font-size":re?`${re}px`:z,"--n-font-family":f,"--n-font-weight-strong":w,"--n-bezier":i,"--n-text-color":u,"--n-mono-3":j,"--n-hue-1":I,"--n-hue-2":G,"--n-hue-3":W,"--n-hue-4":A,"--n-hue-5":$e,"--n-hue-5-2":Ce,"--n-hue-6":ze,"--n-hue-6-2":we,"--n-line-number-text-color":S}}),p=c?Z("code",C(()=>`${e.internalFontSize||"a"}`),$,e):void 0;return{mergedClsPrefix:o,codeRef:t,mergedShowLineNumbers:h,lineNumbers:C(()=>{let i=1;const f=[];let u=!1;for(const z of e.code)z===`
`?(u=!0,f.push(i++)):u=!1;return u||f.push(i++),f.join(`
`)}),cssVars:c?void 0:$,themeClass:p==null?void 0:p.themeClass,onRender:p==null?void 0:p.onRender}},render(){var e,r;const{mergedClsPrefix:a,wordWrap:o,mergedShowLineNumbers:c,onRender:t}=this;return t==null||t(),s("code",{class:[`${a}-code`,this.themeClass,o&&`${a}-code--word-wrap`,c&&`${a}-code--show-line-numbers`],style:this.cssVars,ref:"codeRef"},c?s("pre",{class:`${a}-code__line-numbers`},this.lineNumbers):null,(r=(e=this.$slots).default)===null||r===void 0?void 0:r.call(e))}});function Ze(e){const{textColor3:r,infoColor:a,errorColor:o,successColor:c,warningColor:t,textColor1:l,textColor2:d,railColor:h,fontWeightStrong:b,fontSize:y}=e;return Object.assign(Object.assign({},Se),{contentFontSize:y,titleFontWeight:b,circleBorder:`2px solid ${r}`,circleBorderInfo:`2px solid ${a}`,circleBorderError:`2px solid ${o}`,circleBorderSuccess:`2px solid ${c}`,circleBorderWarning:`2px solid ${t}`,iconColor:r,iconColorInfo:a,iconColorError:o,iconColorSuccess:c,iconColorWarning:t,titleTextColor:l,contentTextColor:d,metaTextColor:r,lineColor:h})}const et={common:ge,self:Ze},tt={success:s(be,null),error:s(ve,null),warning:s(he,null),info:s(pe,null)},rt=q({name:"ProgressCircle",props:{clsPrefix:{type:String,required:!0},status:{type:String,required:!0},strokeWidth:{type:Number,required:!0},fillColor:[String,Object],railColor:String,railStyle:[String,Object],percentage:{type:Number,default:0},offsetDegree:{type:Number,default:0},showIndicator:{type:Boolean,required:!0},indicatorTextColor:String,unit:String,viewBoxWidth:{type:Number,required:!0},gapDegree:{type:Number,required:!0},gapOffsetDegree:{type:Number,default:0}},setup(e,{slots:r}){const a=C(()=>{const t="gradient",{fillColor:l}=e;return typeof l=="object"?`${t}-${Ne(JSON.stringify(l))}`:t});function o(t,l,d,h){const{gapDegree:b,viewBoxWidth:y,strokeWidth:$}=e,p=50,i=0,f=p,u=0,z=2*p,w=50+$/2,S=`M ${w},${w} m ${i},${f}
      a ${p},${p} 0 1 1 ${u},${-z}
      a ${p},${p} 0 1 1 ${-u},${z}`,j=Math.PI*2*p,I={stroke:h==="rail"?d:typeof e.fillColor=="object"?`url(#${a.value})`:d,strokeDasharray:`${Math.min(t,100)/100*(j-b)}px ${y*8}px`,strokeDashoffset:`-${b/2}px`,transformOrigin:l?"center":void 0,transform:l?`rotate(${l}deg)`:void 0};return{pathString:S,pathStyle:I}}const c=()=>{const t=typeof e.fillColor=="object",l=t?e.fillColor.stops[0]:"",d=t?e.fillColor.stops[1]:"";return t&&s("defs",null,s("linearGradient",{id:a.value,x1:"0%",y1:"100%",x2:"100%",y2:"0%"},s("stop",{offset:"0%","stop-color":l}),s("stop",{offset:"100%","stop-color":d})))};return()=>{const{fillColor:t,railColor:l,strokeWidth:d,offsetDegree:h,status:b,percentage:y,showIndicator:$,indicatorTextColor:p,unit:i,gapOffsetDegree:f,clsPrefix:u}=e,{pathString:z,pathStyle:w}=o(100,0,l,"rail"),{pathString:S,pathStyle:j}=o(y,h,t,"fill"),I=100+d;return s("div",{class:`${u}-progress-content`,role:"none"},s("div",{class:`${u}-progress-graph`,"aria-hidden":!0},s("div",{class:`${u}-progress-graph-circle`,style:{transform:f?`rotate(${f}deg)`:void 0}},s("svg",{viewBox:`0 0 ${I} ${I}`},c(),s("g",null,s("path",{class:`${u}-progress-graph-circle-rail`,d:z,"stroke-width":d,"stroke-linecap":"round",fill:"none",style:w})),s("g",null,s("path",{class:[`${u}-progress-graph-circle-fill`,y===0&&`${u}-progress-graph-circle-fill--empty`],d:S,"stroke-width":d,"stroke-linecap":"round",fill:"none",style:j}))))),$?s("div",null,r.default?s("div",{class:`${u}-progress-custom-content`,role:"none"},r.default()):b!=="default"?s("div",{class:`${u}-progress-icon`,"aria-hidden":!0},s(me,{clsPrefix:u},{default:()=>tt[b]})):s("div",{class:`${u}-progress-text`,style:{color:p},role:"none"},s("span",{class:`${u}-progress-text__percentage`},y),s("span",{class:`${u}-progress-text__unit`},i))):null)}}}),nt={success:s(be,null),error:s(ve,null),warning:s(he,null),info:s(pe,null)},it=q({name:"ProgressLine",props:{clsPrefix:{type:String,required:!0},percentage:{type:Number,default:0},railColor:String,railStyle:[String,Object],fillColor:[String,Object],status:{type:String,required:!0},indicatorPlacement:{type:String,required:!0},indicatorTextColor:String,unit:{type:String,default:"%"},processing:{type:Boolean,required:!0},showIndicator:{type:Boolean,required:!0},height:[String,Number],railBorderRadius:[String,Number],fillBorderRadius:[String,Number]},setup(e,{slots:r}){const a=C(()=>O(e.height)),o=C(()=>{var l,d;return typeof e.fillColor=="object"?`linear-gradient(to right, ${(l=e.fillColor)===null||l===void 0?void 0:l.stops[0]} , ${(d=e.fillColor)===null||d===void 0?void 0:d.stops[1]})`:e.fillColor}),c=C(()=>e.railBorderRadius!==void 0?O(e.railBorderRadius):e.height!==void 0?O(e.height,{c:.5}):""),t=C(()=>e.fillBorderRadius!==void 0?O(e.fillBorderRadius):e.railBorderRadius!==void 0?O(e.railBorderRadius):e.height!==void 0?O(e.height,{c:.5}):"");return()=>{const{indicatorPlacement:l,railColor:d,railStyle:h,percentage:b,unit:y,indicatorTextColor:$,status:p,showIndicator:i,processing:f,clsPrefix:u}=e;return s("div",{class:`${u}-progress-content`,role:"none"},s("div",{class:`${u}-progress-graph`,"aria-hidden":!0},s("div",{class:[`${u}-progress-graph-line`,{[`${u}-progress-graph-line--indicator-${l}`]:!0}]},s("div",{class:`${u}-progress-graph-line-rail`,style:[{backgroundColor:d,height:a.value,borderRadius:c.value},h]},s("div",{class:[`${u}-progress-graph-line-fill`,f&&`${u}-progress-graph-line-fill--processing`],style:{maxWidth:`${e.percentage}%`,background:o.value,height:a.value,lineHeight:a.value,borderRadius:t.value}},l==="inside"?s("div",{class:`${u}-progress-graph-line-indicator`,style:{color:$}},r.default?r.default():`${b}${y}`):null)))),i&&l==="outside"?s("div",null,r.default?s("div",{class:`${u}-progress-custom-content`,style:{color:$},role:"none"},r.default()):p==="default"?s("div",{role:"none",class:`${u}-progress-icon ${u}-progress-icon--as-text`,style:{color:$}},b,y):s("div",{class:`${u}-progress-icon`,"aria-hidden":!0},s(me,{clsPrefix:u},{default:()=>nt[p]}))):null)}}});function ce(e,r,a=100){return`m ${a/2} ${a/2-e} a ${e} ${e} 0 1 1 0 ${2*e} a ${e} ${e} 0 1 1 0 -${2*e}`}const ot=q({name:"ProgressMultipleCircle",props:{clsPrefix:{type:String,required:!0},viewBoxWidth:{type:Number,required:!0},percentage:{type:Array,default:[0]},strokeWidth:{type:Number,required:!0},circleGap:{type:Number,required:!0},showIndicator:{type:Boolean,required:!0},fillColor:{type:Array,default:()=>[]},railColor:{type:Array,default:()=>[]},railStyle:{type:Array,default:()=>[]}},setup(e,{slots:r}){const a=C(()=>e.percentage.map((t,l)=>`${Math.PI*t/100*(e.viewBoxWidth/2-e.strokeWidth/2*(1+2*l)-e.circleGap*l)*2}, ${e.viewBoxWidth*8}`)),o=(c,t)=>{const l=e.fillColor[t],d=typeof l=="object"?l.stops[0]:"",h=typeof l=="object"?l.stops[1]:"";return typeof e.fillColor[t]=="object"&&s("linearGradient",{id:`gradient-${t}`,x1:"100%",y1:"0%",x2:"0%",y2:"100%"},s("stop",{offset:"0%","stop-color":d}),s("stop",{offset:"100%","stop-color":h}))};return()=>{const{viewBoxWidth:c,strokeWidth:t,circleGap:l,showIndicator:d,fillColor:h,railColor:b,railStyle:y,percentage:$,clsPrefix:p}=e;return s("div",{class:`${p}-progress-content`,role:"none"},s("div",{class:`${p}-progress-graph`,"aria-hidden":!0},s("div",{class:`${p}-progress-graph-circle`},s("svg",{viewBox:`0 0 ${c} ${c}`},s("defs",null,$.map((i,f)=>o(i,f))),$.map((i,f)=>s("g",{key:f},s("path",{class:`${p}-progress-graph-circle-rail`,d:ce(c/2-t/2*(1+2*f)-l*f,t,c),"stroke-width":t,"stroke-linecap":"round",fill:"none",style:[{strokeDashoffset:0,stroke:b[f]},y[f]]}),s("path",{class:[`${p}-progress-graph-circle-fill`,i===0&&`${p}-progress-graph-circle-fill--empty`],d:ce(c/2-t/2*(1+2*f)-l*f,t,c),"stroke-width":t,"stroke-linecap":"round",fill:"none",style:{strokeDasharray:a.value[f],strokeDashoffset:0,stroke:typeof h[f]=="object"?`url(#gradient-${f})`:h[f]}})))))),d&&r.default?s("div",null,s("div",{class:`${p}-progress-text`},r.default())):null)}}}),lt=R([v("progress",{display:"inline-block"},[v("progress-icon",`
 color: var(--n-icon-color);
 transition: color .3s var(--n-bezier);
 `),P("line",`
 width: 100%;
 display: block;
 `,[v("progress-content",`
 display: flex;
 align-items: center;
 `,[v("progress-graph",{flex:1})]),v("progress-custom-content",{marginLeft:"14px"}),v("progress-icon",`
 width: 30px;
 padding-left: 14px;
 height: var(--n-icon-size-line);
 line-height: var(--n-icon-size-line);
 font-size: var(--n-icon-size-line);
 `,[P("as-text",`
 color: var(--n-text-color-line-outer);
 text-align: center;
 width: 40px;
 font-size: var(--n-font-size);
 padding-left: 4px;
 transition: color .3s var(--n-bezier);
 `)])]),P("circle, dashboard",{width:"120px"},[v("progress-custom-content",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 justify-content: center;
 `),v("progress-text",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 color: inherit;
 font-size: var(--n-font-size-circle);
 color: var(--n-text-color-circle);
 font-weight: var(--n-font-weight-circle);
 transition: color .3s var(--n-bezier);
 white-space: nowrap;
 `),v("progress-icon",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 color: var(--n-icon-color);
 font-size: var(--n-icon-size-circle);
 `)]),P("multiple-circle",`
 width: 200px;
 color: inherit;
 `,[v("progress-text",`
 font-weight: var(--n-font-weight-circle);
 color: var(--n-text-color-circle);
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 justify-content: center;
 transition: color .3s var(--n-bezier);
 `)]),v("progress-content",{position:"relative"}),v("progress-graph",{position:"relative"},[v("progress-graph-circle",[R("svg",{verticalAlign:"bottom"}),v("progress-graph-circle-fill",`
 stroke: var(--n-fill-color);
 transition:
 opacity .3s var(--n-bezier),
 stroke .3s var(--n-bezier),
 stroke-dasharray .3s var(--n-bezier);
 `,[P("empty",{opacity:0})]),v("progress-graph-circle-rail",`
 transition: stroke .3s var(--n-bezier);
 overflow: hidden;
 stroke: var(--n-rail-color);
 `)]),v("progress-graph-line",[P("indicator-inside",[v("progress-graph-line-rail",`
 height: 16px;
 line-height: 16px;
 border-radius: 10px;
 `,[v("progress-graph-line-fill",`
 height: inherit;
 border-radius: 10px;
 `),v("progress-graph-line-indicator",`
 background: #0000;
 white-space: nowrap;
 text-align: right;
 margin-left: 14px;
 margin-right: 14px;
 height: inherit;
 font-size: 12px;
 color: var(--n-text-color-line-inner);
 transition: color .3s var(--n-bezier);
 `)])]),P("indicator-inside-label",`
 height: 16px;
 display: flex;
 align-items: center;
 `,[v("progress-graph-line-rail",`
 flex: 1;
 transition: background-color .3s var(--n-bezier);
 `),v("progress-graph-line-indicator",`
 background: var(--n-fill-color);
 font-size: 12px;
 transform: translateZ(0);
 display: flex;
 vertical-align: middle;
 height: 16px;
 line-height: 16px;
 padding: 0 10px;
 border-radius: 10px;
 position: absolute;
 white-space: nowrap;
 color: var(--n-text-color-line-inner);
 transition:
 right .2s var(--n-bezier),
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `)]),v("progress-graph-line-rail",`
 position: relative;
 overflow: hidden;
 height: var(--n-rail-height);
 border-radius: 5px;
 background-color: var(--n-rail-color);
 transition: background-color .3s var(--n-bezier);
 `,[v("progress-graph-line-fill",`
 background: var(--n-fill-color);
 position: relative;
 border-radius: 5px;
 height: inherit;
 width: 100%;
 max-width: 0%;
 transition:
 background-color .3s var(--n-bezier),
 max-width .2s var(--n-bezier);
 `,[P("processing",[R("&::after",`
 content: "";
 background-image: var(--n-line-bg-processing);
 animation: progress-processing-animation 2s var(--n-bezier) infinite;
 `)])])])])])]),R("@keyframes progress-processing-animation",`
 0% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 100%;
 opacity: 1;
 }
 66% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 0;
 opacity: 0;
 }
 100% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 0;
 opacity: 0;
 }
 `)]),at=Object.assign(Object.assign({},V.props),{processing:Boolean,type:{type:String,default:"line"},gapDegree:Number,gapOffsetDegree:Number,status:{type:String,default:"default"},railColor:[String,Array],railStyle:[String,Array],color:[String,Array,Object],viewBoxWidth:{type:Number,default:100},strokeWidth:{type:Number,default:7},percentage:[Number,Array],unit:{type:String,default:"%"},showIndicator:{type:Boolean,default:!0},indicatorPosition:{type:String,default:"outside"},indicatorPlacement:{type:String,default:"outside"},indicatorTextColor:String,circleGap:{type:Number,default:1},height:Number,borderRadius:[String,Number],fillBorderRadius:[String,Number],offsetDegree:Number}),st=q({name:"Progress",props:at,setup(e){const r=C(()=>e.indicatorPlacement||e.indicatorPosition),a=C(()=>{if(e.gapDegree||e.gapDegree===0)return e.gapDegree;if(e.type==="dashboard")return 75}),{mergedClsPrefixRef:o,inlineThemeDisabled:c}=Y(e),t=V("Progress","-progress",lt,je,e,o),l=C(()=>{const{status:h}=e,{common:{cubicBezierEaseInOut:b},self:{fontSize:y,fontSizeCircle:$,railColor:p,railHeight:i,iconSizeCircle:f,iconSizeLine:u,textColorCircle:z,textColorLineInner:w,textColorLineOuter:S,lineBgProcessing:j,fontWeightCircle:I,[L("iconColor",h)]:G,[L("fillColor",h)]:W}}=t.value;return{"--n-bezier":b,"--n-fill-color":W,"--n-font-size":y,"--n-font-size-circle":$,"--n-font-weight-circle":I,"--n-icon-color":G,"--n-icon-size-circle":f,"--n-icon-size-line":u,"--n-line-bg-processing":j,"--n-rail-color":p,"--n-rail-height":i,"--n-text-color-circle":z,"--n-text-color-line-inner":w,"--n-text-color-line-outer":S}}),d=c?Z("progress",C(()=>e.status[0]),l,e):void 0;return{mergedClsPrefix:o,mergedIndicatorPlacement:r,gapDeg:a,cssVars:c?void 0:l,themeClass:d==null?void 0:d.themeClass,onRender:d==null?void 0:d.onRender}},render(){const{type:e,cssVars:r,indicatorTextColor:a,showIndicator:o,status:c,railColor:t,railStyle:l,color:d,percentage:h,viewBoxWidth:b,strokeWidth:y,mergedIndicatorPlacement:$,unit:p,borderRadius:i,fillBorderRadius:f,height:u,processing:z,circleGap:w,mergedClsPrefix:S,gapDeg:j,gapOffsetDegree:I,themeClass:G,$slots:W,onRender:A}=this;return A==null||A(),s("div",{class:[G,`${S}-progress`,`${S}-progress--${e}`,`${S}-progress--${c}`],style:r,"aria-valuemax":100,"aria-valuemin":0,"aria-valuenow":h,role:e==="circle"||e==="line"||e==="dashboard"?"progressbar":"none"},e==="circle"||e==="dashboard"?s(rt,{clsPrefix:S,status:c,showIndicator:o,indicatorTextColor:a,railColor:t,fillColor:d,railStyle:l,offsetDegree:this.offsetDegree,percentage:h,viewBoxWidth:b,strokeWidth:y,gapDegree:j===void 0?e==="dashboard"?75:0:j,gapOffsetDegree:I,unit:p},W):e==="line"?s(it,{clsPrefix:S,status:c,showIndicator:o,indicatorTextColor:a,railColor:t,fillColor:d,railStyle:l,percentage:h,processing:z,indicatorPlacement:$,unit:p,fillBorderRadius:f,railBorderRadius:i,height:u},W):e==="multiple-circle"?s(ot,{clsPrefix:S,strokeWidth:y,railColor:t,fillColor:d,railStyle:l,viewBoxWidth:b,percentage:h,showIndicator:o,circleGap:w},W):null)}}),ue=1.25,ct=v("timeline",`
 position: relative;
 width: 100%;
 display: flex;
 flex-direction: column;
 line-height: ${ue};
`,[P("horizontal",`
 flex-direction: row;
 `,[R(">",[v("timeline-item",`
 flex-shrink: 0;
 padding-right: 40px;
 `,[P("dashed-line-type",[R(">",[v("timeline-item-timeline",[B("line",`
 background-image: linear-gradient(90deg, var(--n-color-start), var(--n-color-start) 50%, transparent 50%, transparent 100%);
 background-size: 10px 1px;
 `)])])]),R(">",[v("timeline-item-content",`
 margin-top: calc(var(--n-icon-size) + 12px);
 `,[R(">",[B("meta",`
 margin-top: 6px;
 margin-bottom: unset;
 `)])]),v("timeline-item-timeline",`
 width: 100%;
 height: calc(var(--n-icon-size) + 12px);
 `,[B("line",`
 left: var(--n-icon-size);
 top: calc(var(--n-icon-size) / 2 - 1px);
 right: 0px;
 width: unset;
 height: 2px;
 `)])])])])]),P("right-placement",[v("timeline-item",[v("timeline-item-content",`
 text-align: right;
 margin-right: calc(var(--n-icon-size) + 12px);
 `),v("timeline-item-timeline",`
 width: var(--n-icon-size);
 right: 0;
 `)])]),P("left-placement",[v("timeline-item",[v("timeline-item-content",`
 margin-left: calc(var(--n-icon-size) + 12px);
 `),v("timeline-item-timeline",`
 left: 0;
 `)])]),v("timeline-item",`
 position: relative;
 `,[R("&:last-child",[v("timeline-item-timeline",[B("line",`
 display: none;
 `)]),v("timeline-item-content",[B("meta",`
 margin-bottom: 0;
 `)])]),v("timeline-item-content",[B("title",`
 margin: var(--n-title-margin);
 font-size: var(--n-title-font-size);
 transition: color .3s var(--n-bezier);
 font-weight: var(--n-title-font-weight);
 color: var(--n-title-text-color);
 `),B("content",`
 transition: color .3s var(--n-bezier);
 font-size: var(--n-content-font-size);
 color: var(--n-content-text-color);
 `),B("meta",`
 transition: color .3s var(--n-bezier);
 font-size: 12px;
 margin-top: 6px;
 margin-bottom: 20px;
 color: var(--n-meta-text-color);
 `)]),P("dashed-line-type",[v("timeline-item-timeline",[B("line",`
 --n-color-start: var(--n-line-color);
 transition: --n-color-start .3s var(--n-bezier);
 background-color: transparent;
 background-image: linear-gradient(180deg, var(--n-color-start), var(--n-color-start) 50%, transparent 50%, transparent 100%);
 background-size: 1px 10px;
 `)])]),v("timeline-item-timeline",`
 width: calc(var(--n-icon-size) + 12px);
 position: absolute;
 top: calc(var(--n-title-font-size) * ${ue} / 2 - var(--n-icon-size) / 2);
 height: 100%;
 `,[B("circle",`
 border: var(--n-circle-border);
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 width: var(--n-icon-size);
 height: var(--n-icon-size);
 border-radius: var(--n-icon-size);
 box-sizing: border-box;
 `),B("icon",`
 color: var(--n-icon-color);
 font-size: var(--n-icon-size);
 height: var(--n-icon-size);
 width: var(--n-icon-size);
 display: flex;
 align-items: center;
 justify-content: center;
 `),B("line",`
 transition: background-color .3s var(--n-bezier);
 position: absolute;
 top: var(--n-icon-size);
 left: calc(var(--n-icon-size) / 2 - 1px);
 bottom: 0px;
 width: 2px;
 background-color: var(--n-line-color);
 `)])])]),ut=Object.assign(Object.assign({},V.props),{horizontal:Boolean,itemPlacement:{type:String,default:"left"},size:{type:String,default:"medium"},iconSize:Number}),xe=Pe("n-timeline"),dt=q({name:"Timeline",props:ut,setup(e,{slots:r}){const{mergedClsPrefixRef:a}=Y(e),o=V("Timeline","-timeline",ct,et,e,a);return Be(xe,{props:e,mergedThemeRef:o,mergedClsPrefixRef:a}),()=>{const{value:c}=a;return s("div",{class:[`${c}-timeline`,e.horizontal&&`${c}-timeline--horizontal`,`${c}-timeline--${e.size}-size`,!e.horizontal&&`${c}-timeline--${e.itemPlacement}-placement`]},r)}}}),ft={time:[String,Number],title:String,content:String,color:String,lineType:{type:String,default:"default"},type:{type:String,default:"default"}},gt=q({name:"TimelineItem",props:ft,slots:Object,setup(e){const r=fe(xe);r||Re("timeline-item","`n-timeline-item` must be placed inside `n-timeline`."),He();const{inlineThemeDisabled:a}=Y(),o=C(()=>{const{props:{size:t,iconSize:l},mergedThemeRef:d}=r,{type:h}=e,{self:{titleTextColor:b,contentTextColor:y,metaTextColor:$,lineColor:p,titleFontWeight:i,contentFontSize:f,[L("iconSize",t)]:u,[L("titleMargin",t)]:z,[L("titleFontSize",t)]:w,[L("circleBorder",h)]:S,[L("iconColor",h)]:j},common:{cubicBezierEaseInOut:I}}=d.value;return{"--n-bezier":I,"--n-circle-border":S,"--n-icon-color":j,"--n-content-font-size":f,"--n-content-text-color":y,"--n-line-color":p,"--n-meta-text-color":$,"--n-title-font-size":w,"--n-title-font-weight":i,"--n-title-margin":z,"--n-title-text-color":b,"--n-icon-size":O(l)||u}}),c=a?Z("timeline-item",C(()=>{const{props:{size:t,iconSize:l}}=r,{type:d}=e;return`${t[0]}${l||"a"}${d[0]}`}),o,r.props):void 0;return{mergedClsPrefix:r.mergedClsPrefixRef,cssVars:a?void 0:o,themeClass:c==null?void 0:c.themeClass,onRender:c==null?void 0:c.onRender}},render(){const{mergedClsPrefix:e,color:r,onRender:a,$slots:o}=this;return a==null||a(),s("div",{class:[`${e}-timeline-item`,this.themeClass,`${e}-timeline-item--${this.type}-type`,`${e}-timeline-item--${this.lineType}-line-type`],style:this.cssVars},s("div",{class:`${e}-timeline-item-timeline`},s("div",{class:`${e}-timeline-item-timeline__line`}),ie(o.icon,c=>c?s("div",{class:`${e}-timeline-item-timeline__icon`,style:{color:r}},c):s("div",{class:`${e}-timeline-item-timeline__circle`,style:{borderColor:r}}))),s("div",{class:`${e}-timeline-item-content`},ie(o.header,c=>c||this.title?s("div",{class:`${e}-timeline-item-content__title`},c||this.title):null),s("div",{class:`${e}-timeline-item-content__content`},oe(o.default,()=>[this.content])),s("div",{class:`${e}-timeline-item-content__meta`},oe(o.footer,()=>[this.time]))))}}),mt={key:0,class:"text-sm font-semibold"},pt={key:1,class:"text-xs text-gray-400"},ht={class:"text-xs text-center mt-1 text-gray-500"},vt={key:0,class:"text-xs text-center text-gray-400"},bt=q({__name:"QuotaRing",props:{label:{},quota:{},size:{}},setup(e){const r=e,a=C(()=>{var h,b,y;if(typeof((h=r.quota)==null?void 0:h.pct)=="number")return Math.round(r.quota.pct*100);const l=((b=r.quota)==null?void 0:b.total)??0,d=((y=r.quota)==null?void 0:y.remaining)??0;return l?Math.round(d/l*100):null}),o=C(()=>a.value==null?"default":a.value<=10?"error":a.value<=30?"warning":"success"),c=C(()=>{var b;if(!((b=r.quota)!=null&&b.reset_at))return"";const l=new Date(r.quota.reset_at).getTime()-Date.now();if(l<=0)return"";const d=Math.floor(l/36e5),h=Math.floor(l%36e5/6e4);return d>0?`${d}h${h}m`:`${h}m`}),t=C(()=>{var d,h,b;const l=[];return((d=r.quota)==null?void 0:d.remaining)!=null&&((h=r.quota)==null?void 0:h.total)!=null&&l.push(`${r.quota.remaining} / ${r.quota.total}`),(b=r.quota)!=null&&b.reset_at&&l.push(`reset: ${new Date(r.quota.reset_at).toLocaleString()}`),l.join(" · ")||"--"});return(l,d)=>(N(),M(n(Le),{placement:"top"},{trigger:g(()=>[E("div",{class:"quota-ring",style:le({width:(e.size??96)+"px"})},[m(n(st),{type:"circle",percentage:a.value??0,status:o.value,"stroke-width":10,style:le({width:(e.size??96)+"px"})},{default:g(()=>[a.value!=null?(N(),D("span",mt,_(a.value)+"%",1)):(N(),D("span",pt,"--"))]),_:1},8,["percentage","status","style"]),E("div",ht,_(e.label),1),c.value?(N(),D("div",vt,"reset "+_(c.value),1)):F("",!0)],4)]),default:g(()=>[x(" "+_(t.value),1)]),_:1}))}}),yt=(e,r)=>{const a=e.__vccOpts||e;for(const[o,c]of r)a[o]=c;return a},de=yt(bt,[["__scopeId","data-v-036af55d"]]),_t={class:"ml-2 text-xs text-gray-500"},xt={key:0,class:"mt-3 text-right"},$t=q({__name:"EventTimeline",props:{accountId:{}},setup(e){const r=e,{t:a}=ee(),o=T(!1),c=T([]),t=T(0),l=T(1),d=T(20);async function h(){o.value=!0;try{const i=await te.events(r.accountId,l.value,d.value);c.value=i.items,t.value=i.total}catch{}finally{o.value=!1}}Q(h);const b={enabled:"success",disabled:"warning",test_ok:"success",test_fail:"error",refreshed:"info",refresh_failed:"error",cooldown_started:"warning",cooldown_cleared:"info",imported:"info",deleted:"error"};function y(i){return b[i.event_type]??"default"}function $(i){try{return new Date(i).toLocaleString()}catch{return i}}const p=C(()=>i=>{if(i.payload==null)return"";if(typeof i.payload=="string")return i.payload;try{return JSON.stringify(i.payload)}catch{return String(i.payload)}});return(i,f)=>(N(),D("div",null,[m(n(_e),{show:o.value},{default:g(()=>[!o.value&&c.value.length===0?(N(),M(n(Ee),{key:0,description:n(a)("accounts.detail.no_events")},null,8,["description"])):(N(),M(n(dt),{key:1},{default:g(()=>[(N(!0),D(ye,null,Ie(c.value,u=>(N(),M(n(gt),{key:u.id,type:y(u),title:u.event_type,time:$(u.created_at)},{default:g(()=>[m(n(J),{size:"tiny",type:y(u)},{default:g(()=>[x("#"+_(u.id),1)]),_:2},1032,["type"]),E("span",_t,_(p.value(u)),1)]),_:2},1032,["type","title","time"]))),128))]),_:1}))]),_:1},8,["show"]),t.value>d.value?(N(),D("div",xt,[m(n(Me),{page:l.value,"onUpdate:page":[f[0]||(f[0]=u=>l.value=u),h],"page-size":d.value,"item-count":t.value,size:"small"},null,8,["page","page-size","item-count"])])):F("",!0)]))}}),Ct={key:0},zt={class:"text-sm text-gray-500 mb-2"},wt={key:1},kt={class:"text-sm text-gray-500 mb-2"},St=q({__name:"CredentialPeek",props:{accountId:{},accountName:{}},setup(e){const r=e,{t:a}=ee(),o=Te(),c=T(!1),t=T(""),l=T(!1),d=T(null);function h(){c.value=!0,t.value="",d.value=null}async function b(){if(t.value!==r.accountName){o.warning(a("accounts.detail.peek_confirm"));return}l.value=!0;try{const p=await te.peek(r.accountId);d.value=p.credentials}catch{}finally{l.value=!1}}function y(){c.value=!1,d.value=null,t.value=""}const $=p=>p?JSON.stringify(p,null,2):"";return(p,i)=>(N(),D(ye,null,[m(n(H),{size:"small",type:"warning",ghost:"",onClick:h},{default:g(()=>[x(_(n(a)("accounts.detail.peek")),1)]),_:1}),m(n(qe),{show:c.value,"onUpdate:show":i[1]||(i[1]=f=>c.value=f),preset:"card",title:n(a)("accounts.detail.peek"),style:{width:"560px"},"on-after-leave":()=>{d.value=null,t.value=""}},{footer:g(()=>[m(n(X),{justify:"end"},{default:g(()=>[m(n(H),{onClick:y},{default:g(()=>[x(_(n(a)("accounts.add_dialog.cancel")),1)]),_:1}),d.value?F("",!0):(N(),M(n(H),{key:0,type:"primary",loading:l.value,onClick:b},{default:g(()=>[x(_(n(a)("accounts.detail.peek")),1)]),_:1},8,["loading"]))]),_:1})]),default:g(()=>[m(n(X),{vertical:""},{default:g(()=>[m(n(Ve),{type:"warning","show-icon":!0},{default:g(()=>[x(_(n(a)("accounts.detail.peek_warning")),1)]),_:1}),d.value?(N(),D("div",wt,[E("div",kt,_(n(a)("accounts.detail.peek_credentials")),1),m(n(Qe),{code:$(d.value),language:"json","word-wrap":""},null,8,["code"])])):(N(),D("div",Ct,[E("div",zt,_(n(a)("accounts.detail.peek_confirm")),1),m(n(Ge),{value:t.value,"onUpdate:value":i[0]||(i[0]=f=>t.value=f),placeholder:n(a)("accounts.detail.peek_confirm_placeholder")},null,8,["value","placeholder"])]))]),_:1})]),_:1},8,["show","title","on-after-leave"])],64))}}),Nt={key:0,class:"mt-3"},jt={class:"mt-3"},er=q({__name:"Detail",setup(e){const r=We(),a=De(),{t:o}=ee(),c=C(()=>Number(r.params.id)),t=T(null),l=T(!1);async function d(){l.value=!0;try{t.value=await te.get(c.value)}catch{}finally{l.value=!1}}Q(d);const h=Oe(d);function b(i){return i===0?"success":i===1?"default":i===2?"warning":"error"}function y(i){if(!i)return"--";try{return new Date(i).toLocaleString()}catch{return i}}function $(i){return o(i===1?"accounts.detail.refresh_valid_ok":i===2?"accounts.detail.refresh_valid_invalid":"accounts.detail.refresh_valid_unknown")}function p(i){return i===1?"success":i===2?"error":"default"}return(i,f)=>(N(),M(n(_e),{show:l.value},{default:g(()=>[m(n(Ae),{title:t.value?t.value.name:"--",onBack:f[3]||(f[3]=u=>n(a).push("/accounts"))},{extra:g(()=>[m(n(X),null,{default:g(()=>[m(n(H),{size:"small",onClick:f[0]||(f[0]=u=>n(h).doTest(t.value)),disabled:!t.value},{default:g(()=>[x(_(n(o)("accounts.actions.test")),1)]),_:1},8,["disabled"]),m(n(H),{size:"small",onClick:f[1]||(f[1]=u=>n(h).doRefresh(t.value)),disabled:!t.value||t.value.cred_type!=="oauth"&&t.value.cred_type!=="token_pasted"},{default:g(()=>[x(_(n(o)("accounts.actions.refresh")),1)]),_:1},8,["disabled"]),m(n(H),{size:"small",disabled:!t.value||t.value.status!==2,onClick:f[2]||(f[2]=u=>n(h).doClearCooldown(t.value))},{default:g(()=>[x(_(n(o)("accounts.actions.clear_cd")),1)]),_:1},8,["disabled"]),t.value?(N(),M(St,{key:0,"account-id":t.value.id,"account-name":t.value.name},null,8,["account-id","account-name"])):F("",!0)]),_:1})]),_:1},8,["title"]),t.value?(N(),D("div",Nt,[m(n(Fe),{cols:2,"x-gap":12},{default:g(()=>[m(n(ae),null,{default:g(()=>[m(n(U),{size:"small",title:n(o)("accounts.detail.basic")},{default:g(()=>[m(n(se),{column:2,size:"small",bordered:""},{default:g(()=>[m(n(k),{label:n(o)("accounts.columns.id")},{default:g(()=>[x(_(t.value.id),1)]),_:1},8,["label"]),m(n(k),{label:n(o)("accounts.columns.name")},{default:g(()=>[x(_(t.value.name),1)]),_:1},8,["label"]),m(n(k),{label:n(o)("accounts.columns.channel")},{default:g(()=>[x("#"+_(t.value.channel_id),1)]),_:1},8,["label"]),m(n(k),{label:n(o)("accounts.columns.provider")},{default:g(()=>[x(_(t.value.provider),1)]),_:1},8,["label"]),m(n(k),{label:n(o)("accounts.columns.tier")},{default:g(()=>[x(_(t.value.tier),1)]),_:1},8,["label"]),m(n(k),{label:n(o)("accounts.columns.cred_type")},{default:g(()=>[x(_(n(o)(`accounts.cred_type.${t.value.cred_type}`)),1)]),_:1},8,["label"]),m(n(k),{label:"email"},{default:g(()=>[x(_(t.value.email||"--"),1)]),_:1}),m(n(k),{label:n(o)("accounts.columns.status")},{default:g(()=>[m(n(J),{type:b(t.value.status),size:"small"},{default:g(()=>[x(_(n(o)(`accounts.status.${t.value.status}`)),1)]),_:1},8,["type"])]),_:1},8,["label"]),m(n(k),{label:"priority/weight"},{default:g(()=>[x(_(t.value.priority)+" / "+_(t.value.weight),1)]),_:1}),m(n(k),{label:"import_source"},{default:g(()=>[x(_(t.value.import_source||"--"),1)]),_:1}),m(n(k),{label:"external_account_id"},{default:g(()=>[x(_(t.value.external_account_id||"--"),1)]),_:1}),m(n(k),{label:n(o)("accounts.columns.last_used_at")},{default:g(()=>[x(_(y(t.value.last_used_at)),1)]),_:1},8,["label"]),m(n(k),{label:"last_success_at"},{default:g(()=>[x(_(y(t.value.last_success_at)),1)]),_:1}),m(n(k),{label:"last_failure_at"},{default:g(()=>[x(_(y(t.value.last_failure_at)),1)]),_:1}),m(n(k),{label:"cooldown_until"},{default:g(()=>[x(_(y(t.value.cooldown_until)),1)]),_:1}),m(n(k),{label:"created_at"},{default:g(()=>[x(_(y(t.value.created_at)),1)]),_:1}),m(n(k),{label:"updated_at"},{default:g(()=>[x(_(y(t.value.updated_at)),1)]),_:1})]),_:1})]),_:1},8,["title"])]),_:1}),m(n(ae),null,{default:g(()=>[m(n(U),{size:"small",title:n(o)("accounts.detail.token_health")},{default:g(()=>[m(n(se),{column:1,size:"small",bordered:""},{default:g(()=>[m(n(k),{label:n(o)("accounts.detail.token_expires_at")},{default:g(()=>[x(_(y(t.value.access_token_expires_at)),1)]),_:1},8,["label"]),m(n(k),{label:n(o)("accounts.detail.refresh_token_valid")},{default:g(()=>[m(n(J),{type:p(t.value.refresh_token_valid),size:"small"},{default:g(()=>[x(_($(t.value.refresh_token_valid)),1)]),_:1},8,["type"])]),_:1},8,["label"]),m(n(k),{label:n(o)("accounts.detail.consec_failures")},{default:g(()=>[x(_(t.value.consec_failures),1)]),_:1},8,["label"])]),_:1}),E("div",jt,[m(n(X),null,{default:g(()=>[m(de,{label:n(o)("accounts.detail.quota_5h"),quota:t.value.quota_5h},null,8,["label","quota"]),m(de,{label:n(o)("accounts.detail.quota_week"),quota:t.value.quota_week},null,8,["label","quota"])]),_:1})])]),_:1},8,["title"])]),_:1})]),_:1}),m(n(U),{size:"small",title:n(o)("accounts.detail.events"),class:"mt-3"},{default:g(()=>[m($t,{"account-id":t.value.id},null,8,["account-id"])]),_:1},8,["title"])])):F("",!0)]),_:1},8,["show"]))}});export{er as default};
