import{aZ as fe,a3 as C,a4 as ke,ap as ge,x as R,y as v,D as P,A as B,al as q,aU as a,cu as U,bI as Z,cN as Y,cI as V,cJ as Q,bZ as T,cn as ie,ce as Se,d as me,I as pe,W as he,E as ve,k as be,bw as Ne,bT as je,ac as L,ab as Pe,bU as Be,c8 as ne,c6 as oe,cj as Re,bK as N,a7 as M,cs as i,cP as g,ae as x,cm as _,a6 as H,bA as le,ag as m,a9 as D,a8 as G,cw as ee,b as ye,b$ as Ie,cB as Te,B as E,i as qe,f as K,cE as De,cD as We}from"./index-ClKeNBH9.js";import{a as te,u as Oe}from"./useAccountActions-aBbB9753.js";import{f as O}from"./get-48VdzrSm.js";import{N as Le}from"./Tooltip-jxqJeGHW.js";import{_ as Me}from"./_plugin-vue_export-helper-DlAUqK2U.js";import{N as Ee}from"./Pagination-KEuJz1dX.js";import{u as He}from"./use-houdini-BzACH5SR.js";import{N as _e}from"./Spin-sBQV6s4L.js";import{N as Ve}from"./Empty-DQjXSwmQ.js";import{N as J}from"./Tag-BnfjH-Aw.js";import{N as X}from"./Space-C_SQi-yO.js";import{N as Ae}from"./Alert-XMIW4hzt.js";import{a as Fe}from"./Input-C00Z2Txa.js";import{N as Ge}from"./PageHeader-DV_RVema.js";import{a as ae,N as Xe}from"./Grid-DnL_Wmf5.js";import{N as se,a as z}from"./DescriptionsItem-CYvMaqNG.js";import"./Popover-Dm198j0w.js";import"./cssr-DxXR4Bge.js";import"./Select-FWeS_wrt.js";import"./create-DcIarVxf.js";import"./FocusDetector-KCJIcsVz.js";import"./happens-in-CM8LO42l.js";import"./use-locale-CGm--TNI.js";import"./create-ref-setter-C4J8sofl.js";import"./get-slot-Bk_rJcZu.js";function Ue(e,r){const s=fe(ke,null);return C(()=>e.hljs||(s==null?void 0:s.mergedHljsRef.value))}function Ye(e){const{textColor2:r,fontSize:s,fontWeightStrong:l,textColor3:c}=e;return{textColor:r,fontSize:s,fontWeightStrong:l,"mono-3":"#a0a1a7","hue-1":"#0184bb","hue-2":"#4078f2","hue-3":"#a626a4","hue-4":"#50a14f","hue-5":"#e45649","hue-5-2":"#c91243","hue-6":"#986801","hue-6-2":"#c18401",lineNumberTextColor:c}}const Ke={common:ge,self:Ye},Je=R([v("code",`
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
 }`]}]),Ze=Object.assign(Object.assign({},V.props),{language:String,code:{type:String,default:""},trim:{type:Boolean,default:!0},hljs:Object,uri:Boolean,inline:Boolean,wordWrap:Boolean,showLineNumbers:Boolean,internalFontSize:Number,internalNoHighlight:Boolean}),Qe=q({name:"Code",props:Ze,setup(e,{slots:r}){const{internalNoHighlight:s}=e,{mergedClsPrefixRef:l,inlineThemeDisabled:c}=U(),t=T(null),o=s?{value:void 0}:Ue(e),d=(n,f,u)=>{const{value:w}=o;return!w||!(n&&w.getLanguage(n))?null:w.highlight(u?f.trim():f,{language:n}).value},h=C(()=>e.inline||e.wordWrap?!1:e.showLineNumbers),b=()=>{if(r.default)return;const{value:n}=t;if(!n)return;const{language:f}=e,u=e.uri?window.decodeURIComponent(e.code):e.code;if(f){const k=d(f,u,e.trim);if(k!==null){if(e.inline)n.innerHTML=k;else{const S=n.querySelector(".__code__");S&&n.removeChild(S);const j=document.createElement("pre");j.className="__code__",j.innerHTML=k,n.appendChild(j)}return}}if(e.inline){n.textContent=u;return}const w=n.querySelector(".__code__");if(w)w.textContent=u;else{const k=document.createElement("pre");k.className="__code__",k.textContent=u,n.innerHTML="",n.appendChild(k)}};Z(b),Y(ie(e,"language"),b),Y(ie(e,"code"),b),s||Y(o,b);const y=V("Code","-code",Je,Ke,e,l),$=C(()=>{const{common:{cubicBezierEaseInOut:n,fontFamilyMono:f},self:{textColor:u,fontSize:w,fontWeightStrong:k,lineNumberTextColor:S,"mono-3":j,"hue-1":I,"hue-2":A,"hue-3":W,"hue-4":F,"hue-5":$e,"hue-5-2":Ce,"hue-6":we,"hue-6-2":ze}}=y.value,{internalFontSize:re}=e;return{"--n-font-size":re?`${re}px`:w,"--n-font-family":f,"--n-font-weight-strong":k,"--n-bezier":n,"--n-text-color":u,"--n-mono-3":j,"--n-hue-1":I,"--n-hue-2":A,"--n-hue-3":W,"--n-hue-4":F,"--n-hue-5":$e,"--n-hue-5-2":Ce,"--n-hue-6":we,"--n-hue-6-2":ze,"--n-line-number-text-color":S}}),p=c?Q("code",C(()=>`${e.internalFontSize||"a"}`),$,e):void 0;return{mergedClsPrefix:l,codeRef:t,mergedShowLineNumbers:h,lineNumbers:C(()=>{let n=1;const f=[];let u=!1;for(const w of e.code)w===`
`?(u=!0,f.push(n++)):u=!1;return u||f.push(n++),f.join(`
`)}),cssVars:c?void 0:$,themeClass:p==null?void 0:p.themeClass,onRender:p==null?void 0:p.onRender}},render(){var e,r;const{mergedClsPrefix:s,wordWrap:l,mergedShowLineNumbers:c,onRender:t}=this;return t==null||t(),a("code",{class:[`${s}-code`,this.themeClass,l&&`${s}-code--word-wrap`,c&&`${s}-code--show-line-numbers`],style:this.cssVars,ref:"codeRef"},c?a("pre",{class:`${s}-code__line-numbers`},this.lineNumbers):null,(r=(e=this.$slots).default)===null||r===void 0?void 0:r.call(e))}});function et(e){const{textColor3:r,infoColor:s,errorColor:l,successColor:c,warningColor:t,textColor1:o,textColor2:d,railColor:h,fontWeightStrong:b,fontSize:y}=e;return Object.assign(Object.assign({},Se),{contentFontSize:y,titleFontWeight:b,circleBorder:`2px solid ${r}`,circleBorderInfo:`2px solid ${s}`,circleBorderError:`2px solid ${l}`,circleBorderSuccess:`2px solid ${c}`,circleBorderWarning:`2px solid ${t}`,iconColor:r,iconColorInfo:s,iconColorError:l,iconColorSuccess:c,iconColorWarning:t,titleTextColor:o,contentTextColor:d,metaTextColor:r,lineColor:h})}const tt={common:ge,self:et},rt={success:a(be,null),error:a(ve,null),warning:a(he,null),info:a(pe,null)},it=q({name:"ProgressCircle",props:{clsPrefix:{type:String,required:!0},status:{type:String,required:!0},strokeWidth:{type:Number,required:!0},fillColor:[String,Object],railColor:String,railStyle:[String,Object],percentage:{type:Number,default:0},offsetDegree:{type:Number,default:0},showIndicator:{type:Boolean,required:!0},indicatorTextColor:String,unit:String,viewBoxWidth:{type:Number,required:!0},gapDegree:{type:Number,required:!0},gapOffsetDegree:{type:Number,default:0}},setup(e,{slots:r}){const s=C(()=>{const t="gradient",{fillColor:o}=e;return typeof o=="object"?`${t}-${Ne(JSON.stringify(o))}`:t});function l(t,o,d,h){const{gapDegree:b,viewBoxWidth:y,strokeWidth:$}=e,p=50,n=0,f=p,u=0,w=2*p,k=50+$/2,S=`M ${k},${k} m ${n},${f}
      a ${p},${p} 0 1 1 ${u},${-w}
      a ${p},${p} 0 1 1 ${-u},${w}`,j=Math.PI*2*p,I={stroke:h==="rail"?d:typeof e.fillColor=="object"?`url(#${s.value})`:d,strokeDasharray:`${Math.min(t,100)/100*(j-b)}px ${y*8}px`,strokeDashoffset:`-${b/2}px`,transformOrigin:o?"center":void 0,transform:o?`rotate(${o}deg)`:void 0};return{pathString:S,pathStyle:I}}const c=()=>{const t=typeof e.fillColor=="object",o=t?e.fillColor.stops[0]:"",d=t?e.fillColor.stops[1]:"";return t&&a("defs",null,a("linearGradient",{id:s.value,x1:"0%",y1:"100%",x2:"100%",y2:"0%"},a("stop",{offset:"0%","stop-color":o}),a("stop",{offset:"100%","stop-color":d})))};return()=>{const{fillColor:t,railColor:o,strokeWidth:d,offsetDegree:h,status:b,percentage:y,showIndicator:$,indicatorTextColor:p,unit:n,gapOffsetDegree:f,clsPrefix:u}=e,{pathString:w,pathStyle:k}=l(100,0,o,"rail"),{pathString:S,pathStyle:j}=l(y,h,t,"fill"),I=100+d;return a("div",{class:`${u}-progress-content`,role:"none"},a("div",{class:`${u}-progress-graph`,"aria-hidden":!0},a("div",{class:`${u}-progress-graph-circle`,style:{transform:f?`rotate(${f}deg)`:void 0}},a("svg",{viewBox:`0 0 ${I} ${I}`},c(),a("g",null,a("path",{class:`${u}-progress-graph-circle-rail`,d:w,"stroke-width":d,"stroke-linecap":"round",fill:"none",style:k})),a("g",null,a("path",{class:[`${u}-progress-graph-circle-fill`,y===0&&`${u}-progress-graph-circle-fill--empty`],d:S,"stroke-width":d,"stroke-linecap":"round",fill:"none",style:j}))))),$?a("div",null,r.default?a("div",{class:`${u}-progress-custom-content`,role:"none"},r.default()):b!=="default"?a("div",{class:`${u}-progress-icon`,"aria-hidden":!0},a(me,{clsPrefix:u},{default:()=>rt[b]})):a("div",{class:`${u}-progress-text`,style:{color:p},role:"none"},a("span",{class:`${u}-progress-text__percentage`},y),a("span",{class:`${u}-progress-text__unit`},n))):null)}}}),nt={success:a(be,null),error:a(ve,null),warning:a(he,null),info:a(pe,null)},ot=q({name:"ProgressLine",props:{clsPrefix:{type:String,required:!0},percentage:{type:Number,default:0},railColor:String,railStyle:[String,Object],fillColor:[String,Object],status:{type:String,required:!0},indicatorPlacement:{type:String,required:!0},indicatorTextColor:String,unit:{type:String,default:"%"},processing:{type:Boolean,required:!0},showIndicator:{type:Boolean,required:!0},height:[String,Number],railBorderRadius:[String,Number],fillBorderRadius:[String,Number]},setup(e,{slots:r}){const s=C(()=>O(e.height)),l=C(()=>{var o,d;return typeof e.fillColor=="object"?`linear-gradient(to right, ${(o=e.fillColor)===null||o===void 0?void 0:o.stops[0]} , ${(d=e.fillColor)===null||d===void 0?void 0:d.stops[1]})`:e.fillColor}),c=C(()=>e.railBorderRadius!==void 0?O(e.railBorderRadius):e.height!==void 0?O(e.height,{c:.5}):""),t=C(()=>e.fillBorderRadius!==void 0?O(e.fillBorderRadius):e.railBorderRadius!==void 0?O(e.railBorderRadius):e.height!==void 0?O(e.height,{c:.5}):"");return()=>{const{indicatorPlacement:o,railColor:d,railStyle:h,percentage:b,unit:y,indicatorTextColor:$,status:p,showIndicator:n,processing:f,clsPrefix:u}=e;return a("div",{class:`${u}-progress-content`,role:"none"},a("div",{class:`${u}-progress-graph`,"aria-hidden":!0},a("div",{class:[`${u}-progress-graph-line`,{[`${u}-progress-graph-line--indicator-${o}`]:!0}]},a("div",{class:`${u}-progress-graph-line-rail`,style:[{backgroundColor:d,height:s.value,borderRadius:c.value},h]},a("div",{class:[`${u}-progress-graph-line-fill`,f&&`${u}-progress-graph-line-fill--processing`],style:{maxWidth:`${e.percentage}%`,background:l.value,height:s.value,lineHeight:s.value,borderRadius:t.value}},o==="inside"?a("div",{class:`${u}-progress-graph-line-indicator`,style:{color:$}},r.default?r.default():`${b}${y}`):null)))),n&&o==="outside"?a("div",null,r.default?a("div",{class:`${u}-progress-custom-content`,style:{color:$},role:"none"},r.default()):p==="default"?a("div",{role:"none",class:`${u}-progress-icon ${u}-progress-icon--as-text`,style:{color:$}},b,y):a("div",{class:`${u}-progress-icon`,"aria-hidden":!0},a(me,{clsPrefix:u},{default:()=>nt[p]}))):null)}}});function ce(e,r,s=100){return`m ${s/2} ${s/2-e} a ${e} ${e} 0 1 1 0 ${2*e} a ${e} ${e} 0 1 1 0 -${2*e}`}const lt=q({name:"ProgressMultipleCircle",props:{clsPrefix:{type:String,required:!0},viewBoxWidth:{type:Number,required:!0},percentage:{type:Array,default:[0]},strokeWidth:{type:Number,required:!0},circleGap:{type:Number,required:!0},showIndicator:{type:Boolean,required:!0},fillColor:{type:Array,default:()=>[]},railColor:{type:Array,default:()=>[]},railStyle:{type:Array,default:()=>[]}},setup(e,{slots:r}){const s=C(()=>e.percentage.map((t,o)=>`${Math.PI*t/100*(e.viewBoxWidth/2-e.strokeWidth/2*(1+2*o)-e.circleGap*o)*2}, ${e.viewBoxWidth*8}`)),l=(c,t)=>{const o=e.fillColor[t],d=typeof o=="object"?o.stops[0]:"",h=typeof o=="object"?o.stops[1]:"";return typeof e.fillColor[t]=="object"&&a("linearGradient",{id:`gradient-${t}`,x1:"100%",y1:"0%",x2:"0%",y2:"100%"},a("stop",{offset:"0%","stop-color":d}),a("stop",{offset:"100%","stop-color":h}))};return()=>{const{viewBoxWidth:c,strokeWidth:t,circleGap:o,showIndicator:d,fillColor:h,railColor:b,railStyle:y,percentage:$,clsPrefix:p}=e;return a("div",{class:`${p}-progress-content`,role:"none"},a("div",{class:`${p}-progress-graph`,"aria-hidden":!0},a("div",{class:`${p}-progress-graph-circle`},a("svg",{viewBox:`0 0 ${c} ${c}`},a("defs",null,$.map((n,f)=>l(n,f))),$.map((n,f)=>a("g",{key:f},a("path",{class:`${p}-progress-graph-circle-rail`,d:ce(c/2-t/2*(1+2*f)-o*f,t,c),"stroke-width":t,"stroke-linecap":"round",fill:"none",style:[{strokeDashoffset:0,stroke:b[f]},y[f]]}),a("path",{class:[`${p}-progress-graph-circle-fill`,n===0&&`${p}-progress-graph-circle-fill--empty`],d:ce(c/2-t/2*(1+2*f)-o*f,t,c),"stroke-width":t,"stroke-linecap":"round",fill:"none",style:{strokeDasharray:s.value[f],strokeDashoffset:0,stroke:typeof h[f]=="object"?`url(#gradient-${f})`:h[f]}})))))),d&&r.default?a("div",null,a("div",{class:`${p}-progress-text`},r.default())):null)}}}),at=R([v("progress",{display:"inline-block"},[v("progress-icon",`
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
 `)]),st=Object.assign(Object.assign({},V.props),{processing:Boolean,type:{type:String,default:"line"},gapDegree:Number,gapOffsetDegree:Number,status:{type:String,default:"default"},railColor:[String,Array],railStyle:[String,Array],color:[String,Array,Object],viewBoxWidth:{type:Number,default:100},strokeWidth:{type:Number,default:7},percentage:[Number,Array],unit:{type:String,default:"%"},showIndicator:{type:Boolean,default:!0},indicatorPosition:{type:String,default:"outside"},indicatorPlacement:{type:String,default:"outside"},indicatorTextColor:String,circleGap:{type:Number,default:1},height:Number,borderRadius:[String,Number],fillBorderRadius:[String,Number],offsetDegree:Number}),ct=q({name:"Progress",props:st,setup(e){const r=C(()=>e.indicatorPlacement||e.indicatorPosition),s=C(()=>{if(e.gapDegree||e.gapDegree===0)return e.gapDegree;if(e.type==="dashboard")return 75}),{mergedClsPrefixRef:l,inlineThemeDisabled:c}=U(e),t=V("Progress","-progress",at,je,e,l),o=C(()=>{const{status:h}=e,{common:{cubicBezierEaseInOut:b},self:{fontSize:y,fontSizeCircle:$,railColor:p,railHeight:n,iconSizeCircle:f,iconSizeLine:u,textColorCircle:w,textColorLineInner:k,textColorLineOuter:S,lineBgProcessing:j,fontWeightCircle:I,[L("iconColor",h)]:A,[L("fillColor",h)]:W}}=t.value;return{"--n-bezier":b,"--n-fill-color":W,"--n-font-size":y,"--n-font-size-circle":$,"--n-font-weight-circle":I,"--n-icon-color":A,"--n-icon-size-circle":f,"--n-icon-size-line":u,"--n-line-bg-processing":j,"--n-rail-color":p,"--n-rail-height":n,"--n-text-color-circle":w,"--n-text-color-line-inner":k,"--n-text-color-line-outer":S}}),d=c?Q("progress",C(()=>e.status[0]),o,e):void 0;return{mergedClsPrefix:l,mergedIndicatorPlacement:r,gapDeg:s,cssVars:c?void 0:o,themeClass:d==null?void 0:d.themeClass,onRender:d==null?void 0:d.onRender}},render(){const{type:e,cssVars:r,indicatorTextColor:s,showIndicator:l,status:c,railColor:t,railStyle:o,color:d,percentage:h,viewBoxWidth:b,strokeWidth:y,mergedIndicatorPlacement:$,unit:p,borderRadius:n,fillBorderRadius:f,height:u,processing:w,circleGap:k,mergedClsPrefix:S,gapDeg:j,gapOffsetDegree:I,themeClass:A,$slots:W,onRender:F}=this;return F==null||F(),a("div",{class:[A,`${S}-progress`,`${S}-progress--${e}`,`${S}-progress--${c}`],style:r,"aria-valuemax":100,"aria-valuemin":0,"aria-valuenow":h,role:e==="circle"||e==="line"||e==="dashboard"?"progressbar":"none"},e==="circle"||e==="dashboard"?a(it,{clsPrefix:S,status:c,showIndicator:l,indicatorTextColor:s,railColor:t,fillColor:d,railStyle:o,offsetDegree:this.offsetDegree,percentage:h,viewBoxWidth:b,strokeWidth:y,gapDegree:j===void 0?e==="dashboard"?75:0:j,gapOffsetDegree:I,unit:p},W):e==="line"?a(ot,{clsPrefix:S,status:c,showIndicator:l,indicatorTextColor:s,railColor:t,fillColor:d,railStyle:o,percentage:h,processing:w,indicatorPlacement:$,unit:p,fillBorderRadius:f,railBorderRadius:n,height:u},W):e==="multiple-circle"?a(lt,{clsPrefix:S,strokeWidth:y,railColor:t,fillColor:d,railStyle:o,viewBoxWidth:b,percentage:h,showIndicator:l,circleGap:k},W):null)}}),ue=1.25,ut=v("timeline",`
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
 `)])])]),dt=Object.assign(Object.assign({},V.props),{horizontal:Boolean,itemPlacement:{type:String,default:"left"},size:{type:String,default:"medium"},iconSize:Number}),xe=Pe("n-timeline"),ft=q({name:"Timeline",props:dt,setup(e,{slots:r}){const{mergedClsPrefixRef:s}=U(e),l=V("Timeline","-timeline",ut,tt,e,s);return Be(xe,{props:e,mergedThemeRef:l,mergedClsPrefixRef:s}),()=>{const{value:c}=s;return a("div",{class:[`${c}-timeline`,e.horizontal&&`${c}-timeline--horizontal`,`${c}-timeline--${e.size}-size`,!e.horizontal&&`${c}-timeline--${e.itemPlacement}-placement`]},r)}}}),gt={time:[String,Number],title:String,content:String,color:String,lineType:{type:String,default:"default"},type:{type:String,default:"default"}},mt=q({name:"TimelineItem",props:gt,slots:Object,setup(e){const r=fe(xe);r||Re("timeline-item","`n-timeline-item` must be placed inside `n-timeline`."),He();const{inlineThemeDisabled:s}=U(),l=C(()=>{const{props:{size:t,iconSize:o},mergedThemeRef:d}=r,{type:h}=e,{self:{titleTextColor:b,contentTextColor:y,metaTextColor:$,lineColor:p,titleFontWeight:n,contentFontSize:f,[L("iconSize",t)]:u,[L("titleMargin",t)]:w,[L("titleFontSize",t)]:k,[L("circleBorder",h)]:S,[L("iconColor",h)]:j},common:{cubicBezierEaseInOut:I}}=d.value;return{"--n-bezier":I,"--n-circle-border":S,"--n-icon-color":j,"--n-content-font-size":f,"--n-content-text-color":y,"--n-line-color":p,"--n-meta-text-color":$,"--n-title-font-size":k,"--n-title-font-weight":n,"--n-title-margin":w,"--n-title-text-color":b,"--n-icon-size":O(o)||u}}),c=s?Q("timeline-item",C(()=>{const{props:{size:t,iconSize:o}}=r,{type:d}=e;return`${t[0]}${o||"a"}${d[0]}`}),l,r.props):void 0;return{mergedClsPrefix:r.mergedClsPrefixRef,cssVars:s?void 0:l,themeClass:c==null?void 0:c.themeClass,onRender:c==null?void 0:c.onRender}},render(){const{mergedClsPrefix:e,color:r,onRender:s,$slots:l}=this;return s==null||s(),a("div",{class:[`${e}-timeline-item`,this.themeClass,`${e}-timeline-item--${this.type}-type`,`${e}-timeline-item--${this.lineType}-line-type`],style:this.cssVars},a("div",{class:`${e}-timeline-item-timeline`},a("div",{class:`${e}-timeline-item-timeline__line`}),ne(l.icon,c=>c?a("div",{class:`${e}-timeline-item-timeline__icon`,style:{color:r}},c):a("div",{class:`${e}-timeline-item-timeline__circle`,style:{borderColor:r}}))),a("div",{class:`${e}-timeline-item-content`},ne(l.header,c=>c||this.title?a("div",{class:`${e}-timeline-item-content__title`},c||this.title):null),a("div",{class:`${e}-timeline-item-content__content`},oe(l.default,()=>[this.content])),a("div",{class:`${e}-timeline-item-content__meta`},oe(l.footer,()=>[this.time]))))}}),pt={key:0,class:"text-sm font-semibold"},ht={key:1,class:"text-xs text-gray-400"},vt={class:"text-xs text-center mt-1 text-gray-500"},bt={key:0,class:"text-xs text-center text-gray-400"},yt=q({__name:"QuotaRing",props:{label:{},quota:{},size:{}},setup(e){const r=e,s=C(()=>{var h,b,y;if(typeof((h=r.quota)==null?void 0:h.pct)=="number")return Math.round(r.quota.pct*100);const o=((b=r.quota)==null?void 0:b.total)??0,d=((y=r.quota)==null?void 0:y.remaining)??0;return o?Math.round(d/o*100):null}),l=C(()=>s.value==null?"default":s.value<=10?"error":s.value<=30?"warning":"success"),c=C(()=>{var b;if(!((b=r.quota)!=null&&b.reset_at))return"";const o=new Date(r.quota.reset_at).getTime()-Date.now();if(o<=0)return"";const d=Math.floor(o/36e5),h=Math.floor(o%36e5/6e4);return d>0?`${d}h${h}m`:`${h}m`}),t=C(()=>{var d,h,b;const o=[];return((d=r.quota)==null?void 0:d.remaining)!=null&&((h=r.quota)==null?void 0:h.total)!=null&&o.push(`${r.quota.remaining} / ${r.quota.total}`),(b=r.quota)!=null&&b.reset_at&&o.push(`reset: ${new Date(r.quota.reset_at).toLocaleString()}`),o.join(" · ")||"--"});return(o,d)=>(N(),M(i(Le),{placement:"top"},{trigger:g(()=>[H("div",{class:"quota-ring",style:le({width:(e.size??96)+"px"})},[m(i(ct),{type:"circle",percentage:s.value??0,status:l.value,"stroke-width":10,style:le({width:(e.size??96)+"px"})},{default:g(()=>[s.value!=null?(N(),D("span",pt,_(s.value)+"%",1)):(N(),D("span",ht,"--"))]),_:1},8,["percentage","status","style"]),H("div",vt,_(e.label),1),c.value?(N(),D("div",bt,"reset "+_(c.value),1)):G("",!0)],4)]),default:g(()=>[x(" "+_(t.value),1)]),_:1}))}}),de=Me(yt,[["__scopeId","data-v-036af55d"]]),_t={class:"ml-2 text-xs text-gray-500"},xt={key:0,class:"mt-3 text-right"},$t=q({__name:"EventTimeline",props:{accountId:{}},setup(e){const r=e,{t:s}=ee(),l=T(!1),c=T([]),t=T(0),o=T(1),d=T(20);async function h(){l.value=!0;try{const n=await te.events(r.accountId,o.value,d.value);c.value=n.items,t.value=n.total}catch{}finally{l.value=!1}}Z(h);const b={enabled:"success",disabled:"warning",test_ok:"success",test_fail:"error",refreshed:"info",refresh_failed:"error",cooldown_started:"warning",cooldown_cleared:"info",imported:"info",deleted:"error"};function y(n){return b[n.event_type]??"default"}function $(n){try{return new Date(n).toLocaleString()}catch{return n}}const p=C(()=>n=>{if(n.payload==null)return"";if(typeof n.payload=="string")return n.payload;try{return JSON.stringify(n.payload)}catch{return String(n.payload)}});return(n,f)=>(N(),D("div",null,[m(i(_e),{show:l.value},{default:g(()=>[!l.value&&c.value.length===0?(N(),M(i(Ve),{key:0,description:i(s)("accounts.detail.no_events")},null,8,["description"])):(N(),M(i(ft),{key:1},{default:g(()=>[(N(!0),D(ye,null,Ie(c.value,u=>(N(),M(i(mt),{key:u.id,type:y(u),title:u.event_type,time:$(u.created_at)},{default:g(()=>[m(i(J),{size:"tiny",type:y(u)},{default:g(()=>[x("#"+_(u.id),1)]),_:2},1032,["type"]),H("span",_t,_(p.value(u)),1)]),_:2},1032,["type","title","time"]))),128))]),_:1}))]),_:1},8,["show"]),t.value>d.value?(N(),D("div",xt,[m(i(Ee),{page:o.value,"onUpdate:page":[f[0]||(f[0]=u=>o.value=u),h],"page-size":d.value,"item-count":t.value,size:"small"},null,8,["page","page-size","item-count"])])):G("",!0)]))}}),Ct={key:0},wt={class:"text-sm text-gray-500 mb-2"},zt={key:1},kt={class:"text-sm text-gray-500 mb-2"},St=q({__name:"CredentialPeek",props:{accountId:{},accountName:{}},setup(e){const r=e,{t:s}=ee(),l=Te(),c=T(!1),t=T(""),o=T(!1),d=T(null);function h(){c.value=!0,t.value="",d.value=null}async function b(){if(t.value!==r.accountName){l.warning(s("accounts.detail.peek_confirm"));return}o.value=!0;try{const p=await te.peek(r.accountId);d.value=p.credentials}catch{}finally{o.value=!1}}function y(){c.value=!1,d.value=null,t.value=""}const $=p=>p?JSON.stringify(p,null,2):"";return(p,n)=>(N(),D(ye,null,[m(i(E),{size:"small",type:"warning",ghost:"",onClick:h},{default:g(()=>[x(_(i(s)("accounts.detail.peek")),1)]),_:1}),m(i(qe),{show:c.value,"onUpdate:show":n[1]||(n[1]=f=>c.value=f),preset:"card",title:i(s)("accounts.detail.peek"),style:{width:"560px"},"on-after-leave":()=>{d.value=null,t.value=""}},{footer:g(()=>[m(i(X),{justify:"end"},{default:g(()=>[m(i(E),{onClick:y},{default:g(()=>[x(_(i(s)("accounts.add_dialog.cancel")),1)]),_:1}),d.value?G("",!0):(N(),M(i(E),{key:0,type:"primary",loading:o.value,onClick:b},{default:g(()=>[x(_(i(s)("accounts.detail.peek")),1)]),_:1},8,["loading"]))]),_:1})]),default:g(()=>[m(i(X),{vertical:""},{default:g(()=>[m(i(Ae),{type:"warning","show-icon":!0},{default:g(()=>[x(_(i(s)("accounts.detail.peek_warning")),1)]),_:1}),d.value?(N(),D("div",zt,[H("div",kt,_(i(s)("accounts.detail.peek_credentials")),1),m(i(Qe),{code:$(d.value),language:"json","word-wrap":""},null,8,["code"])])):(N(),D("div",Ct,[H("div",wt,_(i(s)("accounts.detail.peek_confirm")),1),m(i(Fe),{value:t.value,"onUpdate:value":n[0]||(n[0]=f=>t.value=f),placeholder:i(s)("accounts.detail.peek_confirm_placeholder")},null,8,["value","placeholder"])]))]),_:1})]),_:1},8,["show","title","on-after-leave"])],64))}}),Nt={key:0,class:"mt-3"},jt={class:"mt-3"},tr=q({__name:"Detail",setup(e){const r=We(),s=De(),{t:l}=ee(),c=C(()=>Number(r.params.id)),t=T(null),o=T(!1);async function d(){o.value=!0;try{t.value=await te.get(c.value)}catch{}finally{o.value=!1}}Z(d);const h=Oe(d);function b(n){return n===0?"success":n===1?"default":n===2?"warning":"error"}function y(n){if(!n)return"--";try{return new Date(n).toLocaleString()}catch{return n}}function $(n){return l(n===1?"accounts.detail.refresh_valid_ok":n===2?"accounts.detail.refresh_valid_invalid":"accounts.detail.refresh_valid_unknown")}function p(n){return n===1?"success":n===2?"error":"default"}return(n,f)=>(N(),M(i(_e),{show:o.value},{default:g(()=>[m(i(Ge),{title:t.value?t.value.name:"--",onBack:f[3]||(f[3]=u=>i(s).push("/accounts"))},{extra:g(()=>[m(i(X),null,{default:g(()=>[m(i(E),{size:"small",onClick:f[0]||(f[0]=u=>i(h).doTest(t.value)),disabled:!t.value},{default:g(()=>[x(_(i(l)("accounts.actions.test")),1)]),_:1},8,["disabled"]),m(i(E),{size:"small",onClick:f[1]||(f[1]=u=>i(h).doRefresh(t.value)),disabled:!t.value||t.value.cred_type!=="oauth"&&t.value.cred_type!=="token_pasted"},{default:g(()=>[x(_(i(l)("accounts.actions.refresh")),1)]),_:1},8,["disabled"]),m(i(E),{size:"small",disabled:!t.value||t.value.status!==2,onClick:f[2]||(f[2]=u=>i(h).doClearCooldown(t.value))},{default:g(()=>[x(_(i(l)("accounts.actions.clear_cd")),1)]),_:1},8,["disabled"]),t.value?(N(),M(St,{key:0,"account-id":t.value.id,"account-name":t.value.name},null,8,["account-id","account-name"])):G("",!0)]),_:1})]),_:1},8,["title"]),t.value?(N(),D("div",Nt,[m(i(Xe),{cols:2,"x-gap":12},{default:g(()=>[m(i(ae),null,{default:g(()=>[m(i(K),{size:"small",title:i(l)("accounts.detail.basic")},{default:g(()=>[m(i(se),{column:2,size:"small",bordered:""},{default:g(()=>[m(i(z),{label:i(l)("accounts.columns.id")},{default:g(()=>[x(_(t.value.id),1)]),_:1},8,["label"]),m(i(z),{label:i(l)("accounts.columns.name")},{default:g(()=>[x(_(t.value.name),1)]),_:1},8,["label"]),m(i(z),{label:i(l)("accounts.columns.channel")},{default:g(()=>[x("#"+_(t.value.channel_id),1)]),_:1},8,["label"]),m(i(z),{label:i(l)("accounts.columns.share_tag")},{default:g(()=>[x(_(t.value.share_tag||"--"),1)]),_:1},8,["label"]),m(i(z),{label:i(l)("accounts.columns.provider")},{default:g(()=>[x(_(t.value.provider),1)]),_:1},8,["label"]),m(i(z),{label:i(l)("accounts.columns.tier")},{default:g(()=>[x(_(t.value.tier),1)]),_:1},8,["label"]),m(i(z),{label:i(l)("accounts.columns.cred_type")},{default:g(()=>[x(_(i(l)(`accounts.cred_type.${t.value.cred_type}`)),1)]),_:1},8,["label"]),m(i(z),{label:"email"},{default:g(()=>[x(_(t.value.email||"--"),1)]),_:1}),m(i(z),{label:i(l)("accounts.columns.status")},{default:g(()=>[m(i(J),{type:b(t.value.status),size:"small"},{default:g(()=>[x(_(i(l)(`accounts.status.${t.value.status}`)),1)]),_:1},8,["type"])]),_:1},8,["label"]),m(i(z),{label:"priority/weight"},{default:g(()=>[x(_(t.value.priority)+" / "+_(t.value.weight),1)]),_:1}),m(i(z),{label:"import_source"},{default:g(()=>[x(_(t.value.import_source||"--"),1)]),_:1}),m(i(z),{label:"external_account_id"},{default:g(()=>[x(_(t.value.external_account_id||"--"),1)]),_:1}),m(i(z),{label:i(l)("accounts.columns.last_used_at")},{default:g(()=>[x(_(y(t.value.last_used_at)),1)]),_:1},8,["label"]),m(i(z),{label:"last_success_at"},{default:g(()=>[x(_(y(t.value.last_success_at)),1)]),_:1}),m(i(z),{label:"last_failure_at"},{default:g(()=>[x(_(y(t.value.last_failure_at)),1)]),_:1}),m(i(z),{label:"cooldown_until"},{default:g(()=>[x(_(y(t.value.cooldown_until)),1)]),_:1}),m(i(z),{label:"created_at"},{default:g(()=>[x(_(y(t.value.created_at)),1)]),_:1}),m(i(z),{label:"updated_at"},{default:g(()=>[x(_(y(t.value.updated_at)),1)]),_:1})]),_:1})]),_:1},8,["title"])]),_:1}),m(i(ae),null,{default:g(()=>[m(i(K),{size:"small",title:i(l)("accounts.detail.token_health")},{default:g(()=>[m(i(se),{column:1,size:"small",bordered:""},{default:g(()=>[m(i(z),{label:i(l)("accounts.detail.token_expires_at")},{default:g(()=>[x(_(y(t.value.access_token_expires_at)),1)]),_:1},8,["label"]),m(i(z),{label:i(l)("accounts.detail.refresh_token_valid")},{default:g(()=>[m(i(J),{type:p(t.value.refresh_token_valid),size:"small"},{default:g(()=>[x(_($(t.value.refresh_token_valid)),1)]),_:1},8,["type"])]),_:1},8,["label"]),m(i(z),{label:i(l)("accounts.detail.consec_failures")},{default:g(()=>[x(_(t.value.consec_failures),1)]),_:1},8,["label"])]),_:1}),H("div",jt,[m(i(X),null,{default:g(()=>[m(de,{label:i(l)("accounts.detail.quota_5h"),quota:t.value.quota_5h},null,8,["label","quota"]),m(de,{label:i(l)("accounts.detail.quota_week"),quota:t.value.quota_week},null,8,["label","quota"])]),_:1})])]),_:1},8,["title"])]),_:1})]),_:1}),m(i(K),{size:"small",title:i(l)("accounts.detail.events"),class:"mt-3"},{default:g(()=>[m($t,{"account-id":t.value.id},null,8,["account-id"])]),_:1},8,["title"])])):G("",!0)]),_:1},8,["show"]))}});export{tr as default};
