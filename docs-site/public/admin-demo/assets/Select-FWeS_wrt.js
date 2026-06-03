import{bI as en,c3 as Ge,cN as be,bG as vn,al as Fe,aU as r,b_ as Oe,aZ as hn,cz as qe,d as Un,m as gn,y as x,A as P,D as U,x as te,G as Ye,aE as bn,c8 as ln,e as Gn,S as qn,c6 as Qn,cu as nn,cF as pn,cI as Se,cJ as tn,bZ as O,a3 as T,cn as V,b1 as Zn,bx as mn,ao as Jn,ac as ge,aQ as Ie,bU as rn,o as Yn,b as Xn,cO as et,b2 as nt,cQ as tt,cL as ot,O as an,cv as lt,bb as it,aS as rt,bn as at,H as ne,cd as st}from"./index-ClKeNBH9.js";import{c as on,i as dt,N as ut,B as ct,a as ft,V as vt,u as Xe}from"./Popover-Dm198j0w.js";import{N as ht}from"./Input-C00Z2Txa.js";import{N as Qe}from"./Tag-BnfjH-Aw.js";import{c as gt,V as sn,a as bt}from"./create-DcIarVxf.js";import{V as pt,b as mt}from"./FocusDetector-KCJIcsVz.js";import{N as wt}from"./Empty-DQjXSwmQ.js";import{h as Be}from"./happens-in-CM8LO42l.js";import{u as dn}from"./get-48VdzrSm.js";import{u as yt}from"./use-locale-CGm--TNI.js";import{u as xt}from"./cssr-DxXR4Bge.js";function wn(e,a){a&&(en(()=>{const{value:d}=e;d&&Ge.registerHandler(d,a)}),be(e,(d,u)=>{u&&Ge.unregisterHandler(u)},{deep:!1}),vn(()=>{const{value:d}=e;d&&Ge.unregisterHandler(d)}))}function un(e){switch(typeof e){case"string":return e||void 0;case"number":return String(e);default:return}}function Ze(e){const a=e.filter(d=>d!==void 0);if(a.length!==0)return a.length===1?a[0]:d=>{e.forEach(u=>{u&&u(d)})}}const Ct=Fe({name:"Checkmark",render(){return r("svg",{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 16 16"},r("g",{fill:"none"},r("path",{d:"M14.046 3.486a.75.75 0 0 1-.032 1.06l-7.93 7.474a.85.85 0 0 1-1.188-.022l-2.68-2.72a.75.75 0 1 1 1.068-1.053l2.234 2.267l7.468-7.038a.75.75 0 0 1 1.06.032z",fill:"currentColor"})))}}),cn=Fe({name:"NBaseSelectGroupHeader",props:{clsPrefix:{type:String,required:!0},tmNode:{type:Object,required:!0}},setup(){const{renderLabelRef:e,renderOptionRef:a,labelFieldRef:d,nodePropsRef:u}=hn(on);return{labelField:d,nodeProps:u,renderLabel:e,renderOption:a}},render(){const{clsPrefix:e,renderLabel:a,renderOption:d,nodeProps:u,tmNode:{rawNode:v}}=this,b=u==null?void 0:u(v),f=a?a(v,!1):Oe(v[this.labelField],v,!1),i=r("div",Object.assign({},b,{class:[`${e}-base-select-group-header`,b==null?void 0:b.class]}),f);return v.render?v.render({node:i,option:v}):d?d({node:i,option:v,selected:!1}):i}});function Ot(e,a){return r(gn,{name:"fade-in-scale-up-transition"},{default:()=>e?r(Un,{clsPrefix:a,class:`${a}-base-select-option__check`},{default:()=>r(Ct)}):null})}const fn=Fe({name:"NBaseSelectOption",props:{clsPrefix:{type:String,required:!0},tmNode:{type:Object,required:!0}},setup(e){const{valueRef:a,pendingTmNodeRef:d,multipleRef:u,valueSetRef:v,renderLabelRef:b,renderOptionRef:f,labelFieldRef:i,valueFieldRef:k,showCheckmarkRef:z,nodePropsRef:_,handleOptionClick:C,handleOptionMouseEnter:$}=hn(on),F=qe(()=>{const{value:R}=d;return R?e.tmNode.key===R.key:!1});function g(R){const{tmNode:S}=e;S.disabled||C(R,S)}function M(R){const{tmNode:S}=e;S.disabled||$(R,S)}function j(R){const{tmNode:S}=e,{value:D}=F;S.disabled||D||$(R,S)}return{multiple:u,isGrouped:qe(()=>{const{tmNode:R}=e,{parent:S}=R;return S&&S.rawNode.type==="group"}),showCheckmark:z,nodeProps:_,isPending:F,isSelected:qe(()=>{const{value:R}=a,{value:S}=u;if(R===null)return!1;const D=e.tmNode.rawNode[k.value];if(S){const{value:W}=v;return W.has(D)}else return R===D}),labelField:i,renderLabel:b,renderOption:f,handleMouseMove:j,handleMouseEnter:M,handleClick:g}},render(){const{clsPrefix:e,tmNode:{rawNode:a},isSelected:d,isPending:u,isGrouped:v,showCheckmark:b,nodeProps:f,renderOption:i,renderLabel:k,handleClick:z,handleMouseEnter:_,handleMouseMove:C}=this,$=Ot(d,e),F=k?[k(a,d),b&&$]:[Oe(a[this.labelField],a,d),b&&$],g=f==null?void 0:f(a),M=r("div",Object.assign({},g,{class:[`${e}-base-select-option`,a.class,g==null?void 0:g.class,{[`${e}-base-select-option--disabled`]:a.disabled,[`${e}-base-select-option--selected`]:d,[`${e}-base-select-option--grouped`]:v,[`${e}-base-select-option--pending`]:u,[`${e}-base-select-option--show-checkmark`]:b}],style:[(g==null?void 0:g.style)||"",a.style||""],onClick:Ze([z,g==null?void 0:g.onClick]),onMouseenter:Ze([_,g==null?void 0:g.onMouseenter]),onMousemove:Ze([C,g==null?void 0:g.onMousemove])}),r("div",{class:`${e}-base-select-option__content`},F));return a.render?a.render({node:M,option:a,selected:d}):i?i({node:M,option:a,selected:d}):M}}),St=x("base-select-menu",`
 line-height: 1.5;
 outline: none;
 z-index: 0;
 position: relative;
 border-radius: var(--n-border-radius);
 transition:
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 background-color: var(--n-color);
`,[x("scrollbar",`
 max-height: var(--n-height);
 `),x("virtual-list",`
 max-height: var(--n-height);
 `),x("base-select-option",`
 min-height: var(--n-option-height);
 font-size: var(--n-option-font-size);
 display: flex;
 align-items: center;
 `,[P("content",`
 z-index: 1;
 white-space: nowrap;
 text-overflow: ellipsis;
 overflow: hidden;
 `)]),x("base-select-group-header",`
 min-height: var(--n-option-height);
 font-size: .93em;
 display: flex;
 align-items: center;
 `),x("base-select-menu-option-wrapper",`
 position: relative;
 width: 100%;
 `),P("loading, empty",`
 display: flex;
 padding: 12px 32px;
 flex: 1;
 justify-content: center;
 `),P("loading",`
 color: var(--n-loading-color);
 font-size: var(--n-loading-size);
 `),P("header",`
 padding: 8px var(--n-option-padding-left);
 font-size: var(--n-option-font-size);
 transition: 
 color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 border-bottom: 1px solid var(--n-action-divider-color);
 color: var(--n-action-text-color);
 `),P("action",`
 padding: 8px var(--n-option-padding-left);
 font-size: var(--n-option-font-size);
 transition: 
 color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 border-top: 1px solid var(--n-action-divider-color);
 color: var(--n-action-text-color);
 `),x("base-select-group-header",`
 position: relative;
 cursor: default;
 padding: var(--n-option-padding);
 color: var(--n-group-header-text-color);
 `),x("base-select-option",`
 cursor: pointer;
 position: relative;
 padding: var(--n-option-padding);
 transition:
 color .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 box-sizing: border-box;
 color: var(--n-option-text-color);
 opacity: 1;
 `,[U("show-checkmark",`
 padding-right: calc(var(--n-option-padding-right) + 20px);
 `),te("&::before",`
 content: "";
 position: absolute;
 left: 4px;
 right: 4px;
 top: 0;
 bottom: 0;
 border-radius: var(--n-border-radius);
 transition: background-color .3s var(--n-bezier);
 `),te("&:active",`
 color: var(--n-option-text-color-pressed);
 `),U("grouped",`
 padding-left: calc(var(--n-option-padding-left) * 1.5);
 `),U("pending",[te("&::before",`
 background-color: var(--n-option-color-pending);
 `)]),U("selected",`
 color: var(--n-option-text-color-active);
 `,[te("&::before",`
 background-color: var(--n-option-color-active);
 `),U("pending",[te("&::before",`
 background-color: var(--n-option-color-active-pending);
 `)])]),U("disabled",`
 cursor: not-allowed;
 `,[Ye("selected",`
 color: var(--n-option-text-color-disabled);
 `),U("selected",`
 opacity: var(--n-option-opacity-disabled);
 `)]),P("check",`
 font-size: 16px;
 position: absolute;
 right: calc(var(--n-option-padding-right) - 4px);
 top: calc(50% - 7px);
 color: var(--n-option-check-color);
 transition: color .3s var(--n-bezier);
 `,[bn({enterScale:"0.5"})])])]),Ft=Fe({name:"InternalSelectMenu",props:Object.assign(Object.assign({},Se.props),{clsPrefix:{type:String,required:!0},scrollable:{type:Boolean,default:!0},treeMate:{type:Object,required:!0},multiple:Boolean,size:{type:String,default:"medium"},value:{type:[String,Number,Array],default:null},autoPending:Boolean,virtualScroll:{type:Boolean,default:!0},show:{type:Boolean,default:!0},labelField:{type:String,default:"label"},valueField:{type:String,default:"value"},loading:Boolean,focusable:Boolean,renderLabel:Function,renderOption:Function,nodeProps:Function,showCheckmark:{type:Boolean,default:!0},onMousedown:Function,onScroll:Function,onFocus:Function,onBlur:Function,onKeyup:Function,onKeydown:Function,onTabOut:Function,onMouseenter:Function,onMouseleave:Function,onResize:Function,resetMenuOnOptionsChange:{type:Boolean,default:!0},inlineThemeDisabled:Boolean,scrollbarProps:Object,onToggle:Function}),setup(e){const{mergedClsPrefixRef:a,mergedRtlRef:d,mergedComponentPropsRef:u}=nn(e),v=pn("InternalSelectMenu",d,a),b=Se("InternalSelectMenu","-internal-select-menu",St,Zn,e,V(e,"clsPrefix")),f=O(null),i=O(null),k=O(null),z=T(()=>e.treeMate.getFlattenedNodes()),_=T(()=>gt(z.value)),C=O(null);function $(){const{treeMate:o}=e;let c=null;const{value:B}=e;B===null?c=o.getFirstAvailableNode():(e.multiple?c=o.getNode((B||[])[(B||[]).length-1]):c=o.getNode(B),(!c||c.disabled)&&(c=o.getFirstAvailableNode())),Y(c||null)}function F(){const{value:o}=C;o&&!e.treeMate.getNode(o.key)&&(C.value=null)}let g;be(()=>e.show,o=>{o?g=be(()=>e.treeMate,()=>{e.resetMenuOnOptionsChange?(e.autoPending?$():F(),mn(le)):F()},{immediate:!0}):g==null||g()},{immediate:!0}),vn(()=>{g==null||g()});const M=T(()=>Jn(b.value.self[ge("optionHeight",e.size)])),j=T(()=>Ie(b.value.self[ge("padding",e.size)])),R=T(()=>e.multiple&&Array.isArray(e.value)?new Set(e.value):new Set),S=T(()=>{const o=z.value;return o&&o.length===0}),D=T(()=>{var o,c;return(c=(o=u==null?void 0:u.value)===null||o===void 0?void 0:o.Select)===null||c===void 0?void 0:c.renderEmpty});function W(o){const{onToggle:c}=e;c&&c(o)}function E(o){const{onScroll:c}=e;c&&c(o)}function I(o){var c;(c=k.value)===null||c===void 0||c.sync(),E(o)}function oe(){var o;(o=k.value)===null||o===void 0||o.sync()}function q(){const{value:o}=C;return o||null}function de(o,c){c.disabled||Y(c,!1)}function pe(o,c){c.disabled||W(c)}function G(o){var c;Be(o,"action")||(c=e.onKeyup)===null||c===void 0||c.call(e,o)}function Q(o){var c;Be(o,"action")||(c=e.onKeydown)===null||c===void 0||c.call(e,o)}function N(o){var c;(c=e.onMousedown)===null||c===void 0||c.call(e,o),!e.focusable&&o.preventDefault()}function ue(){const{value:o}=C;o&&Y(o.getNext({loop:!0}),!0)}function me(){const{value:o}=C;o&&Y(o.getPrev({loop:!0}),!0)}function Y(o,c=!1){C.value=o,c&&le()}function le(){var o,c;const B=C.value;if(!B)return;const X=_.value(B.key);X!==null&&(e.virtualScroll?(o=i.value)===null||o===void 0||o.scrollTo({index:X}):(c=k.value)===null||c===void 0||c.scrollTo({index:X,elSize:M.value}))}function Re(o){var c,B;!((c=f.value)===null||c===void 0)&&c.contains(o.target)&&((B=e.onFocus)===null||B===void 0||B.call(e,o))}function re(o){var c,B;!((c=f.value)===null||c===void 0)&&c.contains(o.relatedTarget)||(B=e.onBlur)===null||B===void 0||B.call(e,o)}rn(on,{handleOptionMouseEnter:de,handleOptionClick:pe,valueSetRef:R,pendingTmNodeRef:C,nodePropsRef:V(e,"nodeProps"),showCheckmarkRef:V(e,"showCheckmark"),multipleRef:V(e,"multiple"),valueRef:V(e,"value"),renderLabelRef:V(e,"renderLabel"),renderOptionRef:V(e,"renderOption"),labelFieldRef:V(e,"labelField"),valueFieldRef:V(e,"valueField")}),rn(dt,f),en(()=>{const{value:o}=k;o&&o.sync()});const ce=T(()=>{const{size:o}=e,{common:{cubicBezierEaseInOut:c},self:{height:B,borderRadius:X,color:we,groupHeaderTextColor:ie,actionDividerColor:H,optionTextColorPressed:ye,optionTextColor:ae,optionTextColorDisabled:Pe,optionTextColorActive:Te,optionOpacityDisabled:Me,optionCheckColor:ve,actionTextColor:he,optionColorPending:ze,optionColorActive:ke,loadingColor:_e,loadingSize:xe,optionColorActivePending:Ce,[ge("optionFontSize",o)]:J,[ge("optionHeight",o)]:t,[ge("optionPadding",o)]:s}}=b.value;return{"--n-height":B,"--n-action-divider-color":H,"--n-action-text-color":he,"--n-bezier":c,"--n-border-radius":X,"--n-color":we,"--n-option-font-size":J,"--n-group-header-text-color":ie,"--n-option-check-color":ve,"--n-option-color-pending":ze,"--n-option-color-active":ke,"--n-option-color-active-pending":Ce,"--n-option-height":t,"--n-option-opacity-disabled":Me,"--n-option-text-color":ae,"--n-option-text-color-active":Te,"--n-option-text-color-disabled":Pe,"--n-option-text-color-pressed":ye,"--n-option-padding":s,"--n-option-padding-left":Ie(s,"left"),"--n-option-padding-right":Ie(s,"right"),"--n-loading-color":_e,"--n-loading-size":xe}}),{inlineThemeDisabled:K}=e,Z=K?tn("internal-select-menu",T(()=>e.size[0]),ce,e):void 0,fe={selfRef:f,next:ue,prev:me,getPendingTmNode:q};return wn(f,e.onResize),Object.assign({mergedTheme:b,mergedClsPrefix:a,rtlEnabled:v,virtualListRef:i,scrollbarRef:k,itemSize:M,padding:j,flattenedNodes:z,empty:S,mergedRenderEmpty:D,virtualListContainer(){const{value:o}=i;return o==null?void 0:o.listElRef},virtualListContent(){const{value:o}=i;return o==null?void 0:o.itemsElRef},doScroll:E,handleFocusin:Re,handleFocusout:re,handleKeyUp:G,handleKeyDown:Q,handleMouseDown:N,handleVirtualListResize:oe,handleVirtualListScroll:I,cssVars:K?void 0:ce,themeClass:Z==null?void 0:Z.themeClass,onRender:Z==null?void 0:Z.onRender},fe)},render(){const{$slots:e,virtualScroll:a,clsPrefix:d,mergedTheme:u,themeClass:v,onRender:b}=this;return b==null||b(),r("div",{ref:"selfRef",tabindex:this.focusable?0:-1,class:[`${d}-base-select-menu`,`${d}-base-select-menu--${this.size}-size`,this.rtlEnabled&&`${d}-base-select-menu--rtl`,v,this.multiple&&`${d}-base-select-menu--multiple`],style:this.cssVars,onFocusin:this.handleFocusin,onFocusout:this.handleFocusout,onKeyup:this.handleKeyUp,onKeydown:this.handleKeyDown,onMousedown:this.handleMouseDown,onMouseenter:this.onMouseenter,onMouseleave:this.onMouseleave},ln(e.header,f=>f&&r("div",{class:`${d}-base-select-menu__header`,"data-header":!0,key:"header"},f)),this.loading?r("div",{class:`${d}-base-select-menu__loading`},r(Gn,{clsPrefix:d,strokeWidth:20})):this.empty?r("div",{class:`${d}-base-select-menu__empty`,"data-empty":!0},Qn(e.empty,()=>{var f;return[((f=this.mergedRenderEmpty)===null||f===void 0?void 0:f.call(this))||r(wt,{theme:u.peers.Empty,themeOverrides:u.peerOverrides.Empty,size:this.size})]})):r(qn,Object.assign({ref:"scrollbarRef",theme:u.peers.Scrollbar,themeOverrides:u.peerOverrides.Scrollbar,scrollable:this.scrollable,container:a?this.virtualListContainer:void 0,content:a?this.virtualListContent:void 0,onScroll:a?void 0:this.doScroll},this.scrollbarProps),{default:()=>a?r(pt,{ref:"virtualListRef",class:`${d}-virtual-list`,items:this.flattenedNodes,itemSize:this.itemSize,showScrollbar:!1,paddingTop:this.padding.top,paddingBottom:this.padding.bottom,onResize:this.handleVirtualListResize,onScroll:this.handleVirtualListScroll,itemResizable:!0},{default:({item:f})=>f.isGroup?r(cn,{key:f.key,clsPrefix:d,tmNode:f}):f.ignored?null:r(fn,{clsPrefix:d,key:f.key,tmNode:f})}):r("div",{class:`${d}-base-select-menu-option-wrapper`,style:{paddingTop:this.padding.top,paddingBottom:this.padding.bottom}},this.flattenedNodes.map(f=>f.isGroup?r(cn,{key:f.key,clsPrefix:d,tmNode:f}):r(fn,{clsPrefix:d,key:f.key,tmNode:f})))}),ln(e.action,f=>f&&[r("div",{class:`${d}-base-select-menu__action`,"data-action":!0,key:"action"},f),r(mt,{onFocus:this.onTabOut,key:"focus-detector"})]))}}),Rt=te([x("base-selection",`
 --n-padding-single: var(--n-padding-single-top) var(--n-padding-single-right) var(--n-padding-single-bottom) var(--n-padding-single-left);
 --n-padding-multiple: var(--n-padding-multiple-top) var(--n-padding-multiple-right) var(--n-padding-multiple-bottom) var(--n-padding-multiple-left);
 position: relative;
 z-index: auto;
 box-shadow: none;
 width: 100%;
 max-width: 100%;
 display: inline-block;
 vertical-align: bottom;
 border-radius: var(--n-border-radius);
 min-height: var(--n-height);
 line-height: 1.5;
 font-size: var(--n-font-size);
 `,[x("base-loading",`
 color: var(--n-loading-color);
 `),x("base-selection-tags","min-height: var(--n-height);"),P("border, state-border",`
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 pointer-events: none;
 border: var(--n-border);
 border-radius: inherit;
 transition:
 box-shadow .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `),P("state-border",`
 z-index: 1;
 border-color: #0000;
 `),x("base-suffix",`
 cursor: pointer;
 position: absolute;
 top: 50%;
 transform: translateY(-50%);
 right: 10px;
 `,[P("arrow",`
 font-size: var(--n-arrow-size);
 color: var(--n-arrow-color);
 transition: color .3s var(--n-bezier);
 `)]),x("base-selection-overlay",`
 display: flex;
 align-items: center;
 white-space: nowrap;
 pointer-events: none;
 position: absolute;
 top: 0;
 right: 0;
 bottom: 0;
 left: 0;
 padding: var(--n-padding-single);
 transition: color .3s var(--n-bezier);
 `,[P("wrapper",`
 flex-basis: 0;
 flex-grow: 1;
 overflow: hidden;
 text-overflow: ellipsis;
 `)]),x("base-selection-placeholder",`
 color: var(--n-placeholder-color);
 `,[P("inner",`
 max-width: 100%;
 overflow: hidden;
 `)]),x("base-selection-tags",`
 cursor: pointer;
 outline: none;
 box-sizing: border-box;
 position: relative;
 z-index: auto;
 display: flex;
 padding: var(--n-padding-multiple);
 flex-wrap: wrap;
 align-items: center;
 width: 100%;
 vertical-align: bottom;
 background-color: var(--n-color);
 border-radius: inherit;
 transition:
 color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `),x("base-selection-label",`
 height: var(--n-height);
 display: inline-flex;
 width: 100%;
 vertical-align: bottom;
 cursor: pointer;
 outline: none;
 z-index: auto;
 box-sizing: border-box;
 position: relative;
 transition:
 color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 border-radius: inherit;
 background-color: var(--n-color);
 align-items: center;
 `,[x("base-selection-input",`
 font-size: inherit;
 line-height: inherit;
 outline: none;
 cursor: pointer;
 box-sizing: border-box;
 border:none;
 width: 100%;
 padding: var(--n-padding-single);
 background-color: #0000;
 color: var(--n-text-color);
 transition: color .3s var(--n-bezier);
 caret-color: var(--n-caret-color);
 `,[P("content",`
 text-overflow: ellipsis;
 overflow: hidden;
 white-space: nowrap; 
 `)]),P("render-label",`
 color: var(--n-text-color);
 `)]),Ye("disabled",[te("&:hover",[P("state-border",`
 box-shadow: var(--n-box-shadow-hover);
 border: var(--n-border-hover);
 `)]),U("focus",[P("state-border",`
 box-shadow: var(--n-box-shadow-focus);
 border: var(--n-border-focus);
 `)]),U("active",[P("state-border",`
 box-shadow: var(--n-box-shadow-active);
 border: var(--n-border-active);
 `),x("base-selection-label","background-color: var(--n-color-active);"),x("base-selection-tags","background-color: var(--n-color-active);")])]),U("disabled","cursor: not-allowed;",[P("arrow",`
 color: var(--n-arrow-color-disabled);
 `),x("base-selection-label",`
 cursor: not-allowed;
 background-color: var(--n-color-disabled);
 `,[x("base-selection-input",`
 cursor: not-allowed;
 color: var(--n-text-color-disabled);
 `),P("render-label",`
 color: var(--n-text-color-disabled);
 `)]),x("base-selection-tags",`
 cursor: not-allowed;
 background-color: var(--n-color-disabled);
 `),x("base-selection-placeholder",`
 cursor: not-allowed;
 color: var(--n-placeholder-color-disabled);
 `)]),x("base-selection-input-tag",`
 height: calc(var(--n-height) - 6px);
 line-height: calc(var(--n-height) - 6px);
 outline: none;
 display: none;
 position: relative;
 margin-bottom: 3px;
 max-width: 100%;
 vertical-align: bottom;
 `,[P("input",`
 font-size: inherit;
 font-family: inherit;
 min-width: 1px;
 padding: 0;
 background-color: #0000;
 outline: none;
 border: none;
 max-width: 100%;
 overflow: hidden;
 width: 1em;
 line-height: inherit;
 cursor: pointer;
 color: var(--n-text-color);
 caret-color: var(--n-caret-color);
 `),P("mirror",`
 position: absolute;
 left: 0;
 top: 0;
 white-space: pre;
 visibility: hidden;
 user-select: none;
 -webkit-user-select: none;
 opacity: 0;
 `)]),["warning","error"].map(e=>U(`${e}-status`,[P("state-border",`border: var(--n-border-${e});`),Ye("disabled",[te("&:hover",[P("state-border",`
 box-shadow: var(--n-box-shadow-hover-${e});
 border: var(--n-border-hover-${e});
 `)]),U("active",[P("state-border",`
 box-shadow: var(--n-box-shadow-active-${e});
 border: var(--n-border-active-${e});
 `),x("base-selection-label",`background-color: var(--n-color-active-${e});`),x("base-selection-tags",`background-color: var(--n-color-active-${e});`)]),U("focus",[P("state-border",`
 box-shadow: var(--n-box-shadow-focus-${e});
 border: var(--n-border-focus-${e});
 `)])])]))]),x("base-selection-popover",`
 margin-bottom: -3px;
 display: flex;
 flex-wrap: wrap;
 margin-right: -8px;
 `),x("base-selection-tag-wrapper",`
 max-width: 100%;
 display: inline-flex;
 padding: 0 7px 3px 0;
 `,[te("&:last-child","padding-right: 0;"),x("tag",`
 font-size: 14px;
 max-width: 100%;
 `,[P("content",`
 line-height: 1.25;
 text-overflow: ellipsis;
 overflow: hidden;
 `)])])]),Pt=Fe({name:"InternalSelection",props:Object.assign(Object.assign({},Se.props),{clsPrefix:{type:String,required:!0},bordered:{type:Boolean,default:void 0},active:Boolean,pattern:{type:String,default:""},placeholder:String,selectedOption:{type:Object,default:null},selectedOptions:{type:Array,default:null},labelField:{type:String,default:"label"},valueField:{type:String,default:"value"},multiple:Boolean,filterable:Boolean,clearable:Boolean,disabled:Boolean,size:{type:String,default:"medium"},loading:Boolean,autofocus:Boolean,showArrow:{type:Boolean,default:!0},inputProps:Object,focused:Boolean,renderTag:Function,onKeydown:Function,onClick:Function,onBlur:Function,onFocus:Function,onDeleteOption:Function,maxTagCount:[String,Number],ellipsisTagPopoverProps:Object,onClear:Function,onPatternInput:Function,onPatternFocus:Function,onPatternBlur:Function,renderLabel:Function,status:String,inlineThemeDisabled:Boolean,ignoreComposition:{type:Boolean,default:!0},onResize:Function}),setup(e){const{mergedClsPrefixRef:a,mergedRtlRef:d}=nn(e),u=pn("InternalSelection",d,a),v=O(null),b=O(null),f=O(null),i=O(null),k=O(null),z=O(null),_=O(null),C=O(null),$=O(null),F=O(null),g=O(!1),M=O(!1),j=O(!1),R=Se("InternalSelection","-internal-selection",Rt,nt,e,V(e,"clsPrefix")),S=T(()=>e.clearable&&!e.disabled&&(j.value||e.active)),D=T(()=>e.selectedOption?e.renderTag?e.renderTag({option:e.selectedOption,handleClose:()=>{}}):e.renderLabel?e.renderLabel(e.selectedOption,!0):Oe(e.selectedOption[e.labelField],e.selectedOption,!0):e.placeholder),W=T(()=>{const t=e.selectedOption;if(t)return t[e.labelField]}),E=T(()=>e.multiple?!!(Array.isArray(e.selectedOptions)&&e.selectedOptions.length):e.selectedOption!==null);function I(){var t;const{value:s}=v;if(s){const{value:A}=b;A&&(A.style.width=`${s.offsetWidth}px`,e.maxTagCount!=="responsive"&&((t=$.value)===null||t===void 0||t.sync({showAllItemsBeforeCalculate:!1})))}}function oe(){const{value:t}=F;t&&(t.style.display="none")}function q(){const{value:t}=F;t&&(t.style.display="inline-block")}be(V(e,"active"),t=>{t||oe()}),be(V(e,"pattern"),()=>{e.multiple&&mn(I)});function de(t){const{onFocus:s}=e;s&&s(t)}function pe(t){const{onBlur:s}=e;s&&s(t)}function G(t){const{onDeleteOption:s}=e;s&&s(t)}function Q(t){const{onClear:s}=e;s&&s(t)}function N(t){const{onPatternInput:s}=e;s&&s(t)}function ue(t){var s;(!t.relatedTarget||!(!((s=f.value)===null||s===void 0)&&s.contains(t.relatedTarget)))&&de(t)}function me(t){var s;!((s=f.value)===null||s===void 0)&&s.contains(t.relatedTarget)||pe(t)}function Y(t){Q(t)}function le(){j.value=!0}function Re(){j.value=!1}function re(t){!e.active||!e.filterable||t.target!==b.value&&t.preventDefault()}function ce(t){G(t)}const K=O(!1);function Z(t){if(t.key==="Backspace"&&!K.value&&!e.pattern.length){const{selectedOptions:s}=e;s!=null&&s.length&&ce(s[s.length-1])}}let fe=null;function o(t){const{value:s}=v;if(s){const A=t.target.value;s.textContent=A,I()}e.ignoreComposition&&K.value?fe=t:N(t)}function c(){K.value=!0}function B(){K.value=!1,e.ignoreComposition&&N(fe),fe=null}function X(t){var s;M.value=!0,(s=e.onPatternFocus)===null||s===void 0||s.call(e,t)}function we(t){var s;M.value=!1,(s=e.onPatternBlur)===null||s===void 0||s.call(e,t)}function ie(){var t,s;if(e.filterable)M.value=!1,(t=z.value)===null||t===void 0||t.blur(),(s=b.value)===null||s===void 0||s.blur();else if(e.multiple){const{value:A}=i;A==null||A.blur()}else{const{value:A}=k;A==null||A.blur()}}function H(){var t,s,A;e.filterable?(M.value=!1,(t=z.value)===null||t===void 0||t.focus()):e.multiple?(s=i.value)===null||s===void 0||s.focus():(A=k.value)===null||A===void 0||A.focus()}function ye(){const{value:t}=b;t&&(q(),t.focus())}function ae(){const{value:t}=b;t&&t.blur()}function Pe(t){const{value:s}=_;s&&s.setTextContent(`+${t}`)}function Te(){const{value:t}=C;return t}function Me(){return b.value}let ve=null;function he(){ve!==null&&window.clearTimeout(ve)}function ze(){e.active||(he(),ve=window.setTimeout(()=>{E.value&&(g.value=!0)},100))}function ke(){he()}function _e(t){t||(he(),g.value=!1)}be(E,t=>{t||(g.value=!1)}),en(()=>{et(()=>{const t=z.value;t&&(e.disabled?t.removeAttribute("tabindex"):t.tabIndex=M.value?-1:0)})}),wn(f,e.onResize);const{inlineThemeDisabled:xe}=e,Ce=T(()=>{const{size:t}=e,{common:{cubicBezierEaseInOut:s},self:{fontWeight:A,borderRadius:je,color:We,placeholderColor:Ke,textColor:$e,paddingSingle:Ee,paddingMultiple:Ne,caretColor:He,colorDisabled:Ue,textColorDisabled:Ae,placeholderColorDisabled:se,colorActive:n,boxShadowFocus:l,boxShadowActive:h,boxShadowHover:w,border:p,borderFocus:m,borderHover:y,borderActive:L,arrowColor:ee,arrowColorDisabled:xn,loadingColor:Cn,colorActiveWarning:On,boxShadowFocusWarning:Sn,boxShadowActiveWarning:Fn,boxShadowHoverWarning:Rn,borderWarning:Pn,borderFocusWarning:Tn,borderHoverWarning:Mn,borderActiveWarning:zn,colorActiveError:kn,boxShadowFocusError:_n,boxShadowActiveError:In,boxShadowHoverError:Bn,borderError:$n,borderFocusError:En,borderHoverError:Nn,borderActiveError:An,clearColor:Dn,clearColorHover:Ln,clearColorPressed:Vn,clearSize:jn,arrowSize:Wn,[ge("height",t)]:Kn,[ge("fontSize",t)]:Hn}}=R.value,De=Ie(Ee),Le=Ie(Ne);return{"--n-bezier":s,"--n-border":p,"--n-border-active":L,"--n-border-focus":m,"--n-border-hover":y,"--n-border-radius":je,"--n-box-shadow-active":h,"--n-box-shadow-focus":l,"--n-box-shadow-hover":w,"--n-caret-color":He,"--n-color":We,"--n-color-active":n,"--n-color-disabled":Ue,"--n-font-size":Hn,"--n-height":Kn,"--n-padding-single-top":De.top,"--n-padding-multiple-top":Le.top,"--n-padding-single-right":De.right,"--n-padding-multiple-right":Le.right,"--n-padding-single-left":De.left,"--n-padding-multiple-left":Le.left,"--n-padding-single-bottom":De.bottom,"--n-padding-multiple-bottom":Le.bottom,"--n-placeholder-color":Ke,"--n-placeholder-color-disabled":se,"--n-text-color":$e,"--n-text-color-disabled":Ae,"--n-arrow-color":ee,"--n-arrow-color-disabled":xn,"--n-loading-color":Cn,"--n-color-active-warning":On,"--n-box-shadow-focus-warning":Sn,"--n-box-shadow-active-warning":Fn,"--n-box-shadow-hover-warning":Rn,"--n-border-warning":Pn,"--n-border-focus-warning":Tn,"--n-border-hover-warning":Mn,"--n-border-active-warning":zn,"--n-color-active-error":kn,"--n-box-shadow-focus-error":_n,"--n-box-shadow-active-error":In,"--n-box-shadow-hover-error":Bn,"--n-border-error":$n,"--n-border-focus-error":En,"--n-border-hover-error":Nn,"--n-border-active-error":An,"--n-clear-size":jn,"--n-clear-color":Dn,"--n-clear-color-hover":Ln,"--n-clear-color-pressed":Vn,"--n-arrow-size":Wn,"--n-font-weight":A}}),J=xe?tn("internal-selection",T(()=>e.size[0]),Ce,e):void 0;return{mergedTheme:R,mergedClearable:S,mergedClsPrefix:a,rtlEnabled:u,patternInputFocused:M,filterablePlaceholder:D,label:W,selected:E,showTagsPanel:g,isComposing:K,counterRef:_,counterWrapperRef:C,patternInputMirrorRef:v,patternInputRef:b,selfRef:f,multipleElRef:i,singleElRef:k,patternInputWrapperRef:z,overflowRef:$,inputTagElRef:F,handleMouseDown:re,handleFocusin:ue,handleClear:Y,handleMouseEnter:le,handleMouseLeave:Re,handleDeleteOption:ce,handlePatternKeyDown:Z,handlePatternInputInput:o,handlePatternInputBlur:we,handlePatternInputFocus:X,handleMouseEnterCounter:ze,handleMouseLeaveCounter:ke,handleFocusout:me,handleCompositionEnd:B,handleCompositionStart:c,onPopoverUpdateShow:_e,focus:H,focusInput:ye,blur:ie,blurInput:ae,updateCounter:Pe,getCounter:Te,getTail:Me,renderLabel:e.renderLabel,cssVars:xe?void 0:Ce,themeClass:J==null?void 0:J.themeClass,onRender:J==null?void 0:J.onRender}},render(){const{status:e,multiple:a,size:d,disabled:u,filterable:v,maxTagCount:b,bordered:f,clsPrefix:i,ellipsisTagPopoverProps:k,onRender:z,renderTag:_,renderLabel:C}=this;z==null||z();const $=b==="responsive",F=typeof b=="number",g=$||F,M=r(Yn,null,{default:()=>r(ht,{clsPrefix:i,loading:this.loading,showArrow:this.showArrow,showClear:this.mergedClearable&&this.selected,onClear:this.handleClear},{default:()=>{var R,S;return(S=(R=this.$slots).arrow)===null||S===void 0?void 0:S.call(R)}})});let j;if(a){const{labelField:R}=this,S=N=>r("div",{class:`${i}-base-selection-tag-wrapper`,key:N.value},_?_({option:N,handleClose:()=>{this.handleDeleteOption(N)}}):r(Qe,{size:d,closable:!N.disabled,disabled:u,onClose:()=>{this.handleDeleteOption(N)},internalCloseIsButtonTag:!1,internalCloseFocusable:!1},{default:()=>C?C(N,!0):Oe(N[R],N,!0)})),D=()=>(F?this.selectedOptions.slice(0,b):this.selectedOptions).map(S),W=v?r("div",{class:`${i}-base-selection-input-tag`,ref:"inputTagElRef",key:"__input-tag__"},r("input",Object.assign({},this.inputProps,{ref:"patternInputRef",tabindex:-1,disabled:u,value:this.pattern,autofocus:this.autofocus,class:`${i}-base-selection-input-tag__input`,onBlur:this.handlePatternInputBlur,onFocus:this.handlePatternInputFocus,onKeydown:this.handlePatternKeyDown,onInput:this.handlePatternInputInput,onCompositionstart:this.handleCompositionStart,onCompositionend:this.handleCompositionEnd})),r("span",{ref:"patternInputMirrorRef",class:`${i}-base-selection-input-tag__mirror`},this.pattern)):null,E=$?()=>r("div",{class:`${i}-base-selection-tag-wrapper`,ref:"counterWrapperRef"},r(Qe,{size:d,ref:"counterRef",onMouseenter:this.handleMouseEnterCounter,onMouseleave:this.handleMouseLeaveCounter,disabled:u})):void 0;let I;if(F){const N=this.selectedOptions.length-b;N>0&&(I=r("div",{class:`${i}-base-selection-tag-wrapper`,key:"__counter__"},r(Qe,{size:d,ref:"counterRef",onMouseenter:this.handleMouseEnterCounter,disabled:u},{default:()=>`+${N}`})))}const oe=$?v?r(sn,{ref:"overflowRef",updateCounter:this.updateCounter,getCounter:this.getCounter,getTail:this.getTail,style:{width:"100%",display:"flex",overflow:"hidden"}},{default:D,counter:E,tail:()=>W}):r(sn,{ref:"overflowRef",updateCounter:this.updateCounter,getCounter:this.getCounter,style:{width:"100%",display:"flex",overflow:"hidden"}},{default:D,counter:E}):F&&I?D().concat(I):D(),q=g?()=>r("div",{class:`${i}-base-selection-popover`},$?D():this.selectedOptions.map(S)):void 0,de=g?Object.assign({show:this.showTagsPanel,trigger:"hover",overlap:!0,placement:"top",width:"trigger",onUpdateShow:this.onPopoverUpdateShow,theme:this.mergedTheme.peers.Popover,themeOverrides:this.mergedTheme.peerOverrides.Popover},k):null,G=(this.selected?!1:this.active?!this.pattern&&!this.isComposing:!0)?r("div",{class:`${i}-base-selection-placeholder ${i}-base-selection-overlay`},r("div",{class:`${i}-base-selection-placeholder__inner`},this.placeholder)):null,Q=v?r("div",{ref:"patternInputWrapperRef",class:`${i}-base-selection-tags`},oe,$?null:W,M):r("div",{ref:"multipleElRef",class:`${i}-base-selection-tags`,tabindex:u?void 0:0},oe,M);j=r(Xn,null,g?r(ut,Object.assign({},de,{scrollable:!0,style:"max-height: calc(var(--v-target-height) * 6.6);"}),{trigger:()=>Q,default:q}):Q,G)}else if(v){const R=this.pattern||this.isComposing,S=this.active?!R:!this.selected,D=this.active?!1:this.selected;j=r("div",{ref:"patternInputWrapperRef",class:`${i}-base-selection-label`,title:this.patternInputFocused?void 0:un(this.label)},r("input",Object.assign({},this.inputProps,{ref:"patternInputRef",class:`${i}-base-selection-input`,value:this.active?this.pattern:"",placeholder:"",readonly:u,disabled:u,tabindex:-1,autofocus:this.autofocus,onFocus:this.handlePatternInputFocus,onBlur:this.handlePatternInputBlur,onInput:this.handlePatternInputInput,onCompositionstart:this.handleCompositionStart,onCompositionend:this.handleCompositionEnd})),D?r("div",{class:`${i}-base-selection-label__render-label ${i}-base-selection-overlay`,key:"input"},r("div",{class:`${i}-base-selection-overlay__wrapper`},_?_({option:this.selectedOption,handleClose:()=>{}}):C?C(this.selectedOption,!0):Oe(this.label,this.selectedOption,!0))):null,S?r("div",{class:`${i}-base-selection-placeholder ${i}-base-selection-overlay`,key:"placeholder"},r("div",{class:`${i}-base-selection-overlay__wrapper`},this.filterablePlaceholder)):null,M)}else j=r("div",{ref:"singleElRef",class:`${i}-base-selection-label`,tabindex:this.disabled?void 0:0},this.label!==void 0?r("div",{class:`${i}-base-selection-input`,title:un(this.label),key:"input"},r("div",{class:`${i}-base-selection-input__content`},_?_({option:this.selectedOption,handleClose:()=>{}}):C?C(this.selectedOption,!0):Oe(this.label,this.selectedOption,!0))):r("div",{class:`${i}-base-selection-placeholder ${i}-base-selection-overlay`,key:"placeholder"},r("div",{class:`${i}-base-selection-placeholder__inner`},this.placeholder)),M);return r("div",{ref:"selfRef",class:[`${i}-base-selection`,this.rtlEnabled&&`${i}-base-selection--rtl`,this.themeClass,e&&`${i}-base-selection--${e}-status`,{[`${i}-base-selection--active`]:this.active,[`${i}-base-selection--selected`]:this.selected||this.active&&this.pattern,[`${i}-base-selection--disabled`]:this.disabled,[`${i}-base-selection--multiple`]:this.multiple,[`${i}-base-selection--focus`]:this.focused}],style:this.cssVars,onClick:this.onClick,onMouseenter:this.handleMouseEnter,onMouseleave:this.handleMouseLeave,onKeydown:this.onKeydown,onFocusin:this.handleFocusin,onFocusout:this.handleFocusout,onMousedown:this.handleMouseDown},j,f?r("div",{class:`${i}-base-selection__border`}):null,f?r("div",{class:`${i}-base-selection__state-border`}):null)}});function Ve(e){return e.type==="group"}function yn(e){return e.type==="ignored"}function Je(e,a){try{return!!(1+a.toString().toLowerCase().indexOf(e.trim().toLowerCase()))}catch{return!1}}function Tt(e,a){return{getIsGroup:Ve,getIgnored:yn,getKey(u){return Ve(u)?u.name||u.key||"key-required":u[e]},getChildren(u){return u[a]}}}function Mt(e,a,d,u){if(!a)return e;function v(b){if(!Array.isArray(b))return[];const f=[];for(const i of b)if(Ve(i)){const k=v(i[u]);k.length&&f.push(Object.assign({},i,{[u]:k}))}else{if(yn(i))continue;a(d,i)&&f.push(i)}return f}return v(e)}function zt(e,a,d){const u=new Map;return e.forEach(v=>{Ve(v)?v[d].forEach(b=>{u.set(b[a],b)}):u.set(v[a],v)}),u}const kt=te([x("select",`
 z-index: auto;
 outline: none;
 width: 100%;
 position: relative;
 font-weight: var(--n-font-weight);
 `),x("select-menu",`
 margin: 4px 0;
 box-shadow: var(--n-menu-box-shadow);
 `,[bn({originalTransition:"background-color .3s var(--n-bezier), box-shadow .3s var(--n-bezier)"})])]),_t=Object.assign(Object.assign({},Se.props),{to:Xe.propTo,bordered:{type:Boolean,default:void 0},clearable:Boolean,clearCreatedOptionsOnClear:{type:Boolean,default:!0},clearFilterAfterSelect:{type:Boolean,default:!0},options:{type:Array,default:()=>[]},defaultValue:{type:[String,Number,Array],default:null},keyboard:{type:Boolean,default:!0},value:[String,Number,Array],placeholder:String,menuProps:Object,multiple:Boolean,size:String,menuSize:{type:String},filterable:Boolean,disabled:{type:Boolean,default:void 0},remote:Boolean,loading:Boolean,filter:Function,placement:{type:String,default:"bottom-start"},widthMode:{type:String,default:"trigger"},tag:Boolean,onCreate:Function,fallbackOption:{type:[Function,Boolean],default:void 0},show:{type:Boolean,default:void 0},showArrow:{type:Boolean,default:!0},maxTagCount:[Number,String],ellipsisTagPopoverProps:Object,consistentMenuWidth:{type:Boolean,default:!0},virtualScroll:{type:Boolean,default:!0},labelField:{type:String,default:"label"},valueField:{type:String,default:"value"},childrenField:{type:String,default:"children"},renderLabel:Function,renderOption:Function,renderTag:Function,"onUpdate:value":[Function,Array],inputProps:Object,nodeProps:Function,ignoreComposition:{type:Boolean,default:!0},showOnFocus:Boolean,onUpdateValue:[Function,Array],onBlur:[Function,Array],onClear:[Function,Array],onFocus:[Function,Array],onScroll:[Function,Array],onSearch:[Function,Array],onUpdateShow:[Function,Array],"onUpdate:show":[Function,Array],displayDirective:{type:String,default:"show"},resetMenuOnOptionsChange:{type:Boolean,default:!0},status:String,showCheckmark:{type:Boolean,default:!0},scrollbarProps:Object,onChange:[Function,Array],items:Array}),Kt=Fe({name:"Select",props:_t,slots:Object,setup(e){const{mergedClsPrefixRef:a,mergedBorderedRef:d,namespaceRef:u,inlineThemeDisabled:v,mergedComponentPropsRef:b}=nn(e),f=Se("Select","-select",kt,st,e,a),i=O(e.defaultValue),k=V(e,"value"),z=dn(k,i),_=O(!1),C=O(""),$=xt(e,["items","options"]),F=O([]),g=O([]),M=T(()=>g.value.concat(F.value).concat($.value)),j=T(()=>{const{filter:n}=e;if(n)return n;const{labelField:l,valueField:h}=e;return(w,p)=>{if(!p)return!1;const m=p[l];if(typeof m=="string")return Je(w,m);const y=p[h];return typeof y=="string"?Je(w,y):typeof y=="number"?Je(w,String(y)):!1}}),R=T(()=>{if(e.remote)return $.value;{const{value:n}=M,{value:l}=C;return!l.length||!e.filterable?n:Mt(n,j.value,l,e.childrenField)}}),S=T(()=>{const{valueField:n,childrenField:l}=e,h=Tt(n,l);return bt(R.value,h)}),D=T(()=>zt(M.value,e.valueField,e.childrenField)),W=O(!1),E=dn(V(e,"show"),W),I=O(null),oe=O(null),q=O(null),{localeRef:de}=yt("Select"),pe=T(()=>{var n;return(n=e.placeholder)!==null&&n!==void 0?n:de.value.placeholder}),G=[],Q=O(new Map),N=T(()=>{const{fallbackOption:n}=e;if(n===void 0){const{labelField:l,valueField:h}=e;return w=>({[l]:String(w),[h]:w})}return n===!1?!1:l=>Object.assign(n(l),{value:l})});function ue(n){const l=e.remote,{value:h}=Q,{value:w}=D,{value:p}=N,m=[];return n.forEach(y=>{if(w.has(y))m.push(w.get(y));else if(l&&h.has(y))m.push(h.get(y));else if(p){const L=p(y);L&&m.push(L)}}),m}const me=T(()=>{if(e.multiple){const{value:n}=z;return Array.isArray(n)?ue(n):[]}return null}),Y=T(()=>{const{value:n}=z;return!e.multiple&&!Array.isArray(n)?n===null?null:ue([n])[0]||null:null}),le=lt(e,{mergedSize:n=>{var l,h;const{size:w}=e;if(w)return w;const{mergedSize:p}=n||{};if(p!=null&&p.value)return p.value;const m=(h=(l=b==null?void 0:b.value)===null||l===void 0?void 0:l.Select)===null||h===void 0?void 0:h.size;return m||"medium"}}),{mergedSizeRef:Re,mergedDisabledRef:re,mergedStatusRef:ce}=le;function K(n,l){const{onChange:h,"onUpdate:value":w,onUpdateValue:p}=e,{nTriggerFormChange:m,nTriggerFormInput:y}=le;h&&ne(h,n,l),p&&ne(p,n,l),w&&ne(w,n,l),i.value=n,m(),y()}function Z(n){const{onBlur:l}=e,{nTriggerFormBlur:h}=le;l&&ne(l,n),h()}function fe(){const{onClear:n}=e;n&&ne(n)}function o(n){const{onFocus:l,showOnFocus:h}=e,{nTriggerFormFocus:w}=le;l&&ne(l,n),w(),h&&ie()}function c(n){const{onSearch:l}=e;l&&ne(l,n)}function B(n){const{onScroll:l}=e;l&&ne(l,n)}function X(){var n;const{remote:l,multiple:h}=e;if(l){const{value:w}=Q;if(h){const{valueField:p}=e;(n=me.value)===null||n===void 0||n.forEach(m=>{w.set(m[p],m)})}else{const p=Y.value;p&&w.set(p[e.valueField],p)}}}function we(n){const{onUpdateShow:l,"onUpdate:show":h}=e;l&&ne(l,n),h&&ne(h,n),W.value=n}function ie(){re.value||(we(!0),W.value=!0,e.filterable&&Ne())}function H(){we(!1)}function ye(){C.value="",g.value=G}const ae=O(!1);function Pe(){e.filterable&&(ae.value=!0)}function Te(){e.filterable&&(ae.value=!1,E.value||ye())}function Me(){re.value||(E.value?e.filterable?Ne():H():ie())}function ve(n){var l,h;!((h=(l=q.value)===null||l===void 0?void 0:l.selfRef)===null||h===void 0)&&h.contains(n.relatedTarget)||(_.value=!1,Z(n),H())}function he(n){o(n),_.value=!0}function ze(){_.value=!0}function ke(n){var l;!((l=I.value)===null||l===void 0)&&l.$el.contains(n.relatedTarget)||(_.value=!1,Z(n),H())}function _e(){var n;(n=I.value)===null||n===void 0||n.focus(),H()}function xe(n){var l;E.value&&(!((l=I.value)===null||l===void 0)&&l.$el.contains(rt(n))||H())}function Ce(n){if(!Array.isArray(n))return[];if(N.value)return Array.from(n);{const{remote:l}=e,{value:h}=D;if(l){const{value:w}=Q;return n.filter(p=>h.has(p)||w.has(p))}else return n.filter(w=>h.has(w))}}function J(n){t(n.rawNode)}function t(n){if(re.value)return;const{tag:l,remote:h,clearFilterAfterSelect:w,valueField:p}=e;if(l&&!h){const{value:m}=g,y=m[0]||null;if(y){const L=F.value;L.length?L.push(y):F.value=[y],g.value=G}}if(h&&Q.value.set(n[p],n),e.multiple){const m=Ce(z.value),y=m.findIndex(L=>L===n[p]);if(~y){if(m.splice(y,1),l&&!h){const L=s(n[p]);~L&&(F.value.splice(L,1),w&&(C.value=""))}}else m.push(n[p]),w&&(C.value="");K(m,ue(m))}else{if(l&&!h){const m=s(n[p]);~m?F.value=[F.value[m]]:F.value=G}Ee(),H(),K(n[p],n)}}function s(n){return F.value.findIndex(h=>h[e.valueField]===n)}function A(n){E.value||ie();const{value:l}=n.target;C.value=l;const{tag:h,remote:w}=e;if(c(l),h&&!w){if(!l){g.value=G;return}const{onCreate:p}=e,m=p?p(l):{[e.labelField]:l,[e.valueField]:l},{valueField:y,labelField:L}=e;$.value.some(ee=>ee[y]===m[y]||ee[L]===m[L])||F.value.some(ee=>ee[y]===m[y]||ee[L]===m[L])?g.value=G:g.value=[m]}}function je(n){n.stopPropagation();const{multiple:l,tag:h,remote:w,clearCreatedOptionsOnClear:p}=e;!l&&e.filterable&&H(),h&&!w&&p&&(F.value=G),fe(),l?K([],[]):K(null,null)}function We(n){!Be(n,"action")&&!Be(n,"empty")&&!Be(n,"header")&&n.preventDefault()}function Ke(n){B(n)}function $e(n){var l,h,w,p,m;if(!e.keyboard){n.preventDefault();return}switch(n.key){case" ":if(e.filterable)break;n.preventDefault();case"Enter":if(!(!((l=I.value)===null||l===void 0)&&l.isComposing)){if(E.value){const y=(h=q.value)===null||h===void 0?void 0:h.getPendingTmNode();y?J(y):e.filterable||(H(),Ee())}else if(ie(),e.tag&&ae.value){const y=g.value[0];if(y){const L=y[e.valueField],{value:ee}=z;e.multiple&&Array.isArray(ee)&&ee.includes(L)||t(y)}}}n.preventDefault();break;case"ArrowUp":if(n.preventDefault(),e.loading)return;E.value&&((w=q.value)===null||w===void 0||w.prev());break;case"ArrowDown":if(n.preventDefault(),e.loading)return;E.value?(p=q.value)===null||p===void 0||p.next():ie();break;case"Escape":E.value&&(at(n),H()),(m=I.value)===null||m===void 0||m.focus();break}}function Ee(){var n;(n=I.value)===null||n===void 0||n.focus()}function Ne(){var n;(n=I.value)===null||n===void 0||n.focusInput()}function He(){var n;E.value&&((n=oe.value)===null||n===void 0||n.syncPosition())}X(),be(V(e,"options"),X);const Ue={focus:()=>{var n;(n=I.value)===null||n===void 0||n.focus()},focusInput:()=>{var n;(n=I.value)===null||n===void 0||n.focusInput()},blur:()=>{var n;(n=I.value)===null||n===void 0||n.blur()},blurInput:()=>{var n;(n=I.value)===null||n===void 0||n.blurInput()}},Ae=T(()=>{const{self:{menuBoxShadow:n}}=f.value;return{"--n-menu-box-shadow":n}}),se=v?tn("select",void 0,Ae,e):void 0;return Object.assign(Object.assign({},Ue),{mergedStatus:ce,mergedClsPrefix:a,mergedBordered:d,namespace:u,treeMate:S,isMounted:it(),triggerRef:I,menuRef:q,pattern:C,uncontrolledShow:W,mergedShow:E,adjustedTo:Xe(e),uncontrolledValue:i,mergedValue:z,followerRef:oe,localizedPlaceholder:pe,selectedOption:Y,selectedOptions:me,mergedSize:Re,mergedDisabled:re,focused:_,activeWithoutMenuOpen:ae,inlineThemeDisabled:v,onTriggerInputFocus:Pe,onTriggerInputBlur:Te,handleTriggerOrMenuResize:He,handleMenuFocus:ze,handleMenuBlur:ke,handleMenuTabOut:_e,handleTriggerClick:Me,handleToggle:J,handleDeleteOption:t,handlePatternInput:A,handleClear:je,handleTriggerBlur:ve,handleTriggerFocus:he,handleKeydown:$e,handleMenuAfterLeave:ye,handleMenuClickOutside:xe,handleMenuScroll:Ke,handleMenuKeydown:$e,handleMenuMousedown:We,mergedTheme:f,cssVars:v?void 0:Ae,themeClass:se==null?void 0:se.themeClass,onRender:se==null?void 0:se.onRender})},render(){return r("div",{class:`${this.mergedClsPrefix}-select`},r(ct,null,{default:()=>[r(ft,null,{default:()=>r(Pt,{ref:"triggerRef",inlineThemeDisabled:this.inlineThemeDisabled,status:this.mergedStatus,inputProps:this.inputProps,clsPrefix:this.mergedClsPrefix,showArrow:this.showArrow,maxTagCount:this.maxTagCount,ellipsisTagPopoverProps:this.ellipsisTagPopoverProps,bordered:this.mergedBordered,active:this.activeWithoutMenuOpen||this.mergedShow,pattern:this.pattern,placeholder:this.localizedPlaceholder,selectedOption:this.selectedOption,selectedOptions:this.selectedOptions,multiple:this.multiple,renderTag:this.renderTag,renderLabel:this.renderLabel,filterable:this.filterable,clearable:this.clearable,disabled:this.mergedDisabled,size:this.mergedSize,theme:this.mergedTheme.peers.InternalSelection,labelField:this.labelField,valueField:this.valueField,themeOverrides:this.mergedTheme.peerOverrides.InternalSelection,loading:this.loading,focused:this.focused,onClick:this.handleTriggerClick,onDeleteOption:this.handleDeleteOption,onPatternInput:this.handlePatternInput,onClear:this.handleClear,onBlur:this.handleTriggerBlur,onFocus:this.handleTriggerFocus,onKeydown:this.handleKeydown,onPatternBlur:this.onTriggerInputBlur,onPatternFocus:this.onTriggerInputFocus,onResize:this.handleTriggerOrMenuResize,ignoreComposition:this.ignoreComposition},{arrow:()=>{var e,a;return[(a=(e=this.$slots).arrow)===null||a===void 0?void 0:a.call(e)]}})}),r(vt,{ref:"followerRef",show:this.mergedShow,to:this.adjustedTo,teleportDisabled:this.adjustedTo===Xe.tdkey,containerClass:this.namespace,width:this.consistentMenuWidth?"target":void 0,minWidth:"target",placement:this.placement},{default:()=>r(gn,{name:"fade-in-scale-up-transition",appear:this.isMounted,onAfterLeave:this.handleMenuAfterLeave},{default:()=>{var e,a,d;return this.mergedShow||this.displayDirective==="show"?((e=this.onRender)===null||e===void 0||e.call(this),tt(r(Ft,Object.assign({},this.menuProps,{ref:"menuRef",onResize:this.handleTriggerOrMenuResize,inlineThemeDisabled:this.inlineThemeDisabled,virtualScroll:this.consistentMenuWidth&&this.virtualScroll,class:[`${this.mergedClsPrefix}-select-menu`,this.themeClass,(a=this.menuProps)===null||a===void 0?void 0:a.class],clsPrefix:this.mergedClsPrefix,focusable:!0,labelField:this.labelField,valueField:this.valueField,autoPending:!0,nodeProps:this.nodeProps,theme:this.mergedTheme.peers.InternalSelectMenu,themeOverrides:this.mergedTheme.peerOverrides.InternalSelectMenu,treeMate:this.treeMate,multiple:this.multiple,size:this.menuSize,renderOption:this.renderOption,renderLabel:this.renderLabel,value:this.mergedValue,style:[(d=this.menuProps)===null||d===void 0?void 0:d.style,this.cssVars],onToggle:this.handleToggle,onScroll:this.handleMenuScroll,onFocus:this.handleMenuFocus,onBlur:this.handleMenuBlur,onKeydown:this.handleMenuKeydown,onTabOut:this.handleMenuTabOut,onMousedown:this.handleMenuMousedown,show:this.mergedShow,showCheckmark:this.showCheckmark,resetMenuOnOptionsChange:this.resetMenuOnOptionsChange,scrollbarProps:this.scrollbarProps}),{empty:()=>{var u,v;return[(v=(u=this.$slots).empty)===null||v===void 0?void 0:v.call(u)]},header:()=>{var u,v;return[(v=(u=this.$slots).header)===null||v===void 0?void 0:v.call(u)]},action:()=>{var u,v;return[(v=(u=this.$slots).action)===null||v===void 0?void 0:v.call(u)]}}),this.displayDirective==="show"?[[ot,this.mergedShow],[an,this.handleMenuClickOutside,void 0,{capture:!0}]]:[[an,this.handleMenuClickOutside,void 0,{capture:!0}]])):null}})})]}))}});export{Ft as N,Kt as a,Tt as c,Ze as m};
