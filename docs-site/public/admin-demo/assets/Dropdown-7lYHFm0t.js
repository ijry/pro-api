import{B as Ne,a as Re,V as Pe,r as ke,N as Ke,p as ae}from"./Popover-BQ0-iUw6.js";import{bX as T,cL as se,ak as z,aS as d,aa as re,bY as W,aX as B,bn as ue,m as Ce,cx as G,a2 as b,bN as ce,bS as H,b as Ie,cK as _e,p as Oe,bt as ze,aq as De,y as x,aC as $e,x as M,G as ie,D as k,A as O,bi as Ae,cs as Fe,cG as pe,cH as Te,at as Be,H as oe,cl as K,ab as F}from"./index-BMyk45kF.js";import{N as je}from"./Icon-Cx7n14g8.js";import{h as de}from"./happens-in-CM8LO42l.js";import{u as Le}from"./get-D86kArqK.js";import{u as Me}from"./use-keyboard-B03E-gRL.js";import{c as He}from"./create-ref-setter-C4J8sofl.js";import{a as Ee}from"./create-BM_mgYTJ.js";function Ue(e,i,a){const t=T(e.value);let r=null;return se(e,n=>{r!==null&&window.clearTimeout(r),n===!0?a&&!a.value?t.value=!0:r=window.setTimeout(()=>{t.value=!0},i):t.value=!1}),t}const qe=z({name:"ChevronRight",render(){return d("svg",{viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg"},d("path",{d:"M5.64645 3.14645C5.45118 3.34171 5.45118 3.65829 5.64645 3.85355L9.79289 8L5.64645 12.1464C5.45118 12.3417 5.45118 12.6583 5.64645 12.8536C5.84171 13.0488 6.15829 13.0488 6.35355 12.8536L10.8536 8.35355C11.0488 8.15829 11.0488 7.84171 10.8536 7.64645L6.35355 3.14645C6.15829 2.95118 5.84171 2.95118 5.64645 3.14645Z",fill:"currentColor"}))}}),te=re("n-dropdown-menu"),V=re("n-dropdown"),le=re("n-dropdown-option"),fe=z({name:"DropdownDivider",props:{clsPrefix:{type:String,required:!0}},render(){return d("div",{class:`${this.clsPrefix}-dropdown-divider`})}}),Ge=z({name:"DropdownGroupHeader",props:{clsPrefix:{type:String,required:!0},tmNode:{type:Object,required:!0}},setup(){const{showIconRef:e,hasSubmenuRef:i}=B(te),{renderLabelRef:a,labelFieldRef:t,nodePropsRef:r,renderOptionRef:n}=B(V);return{labelField:t,showIcon:e,hasSubmenu:i,renderLabel:a,nodeProps:r,renderOption:n}},render(){var e;const{clsPrefix:i,hasSubmenu:a,showIcon:t,nodeProps:r,renderLabel:n,renderOption:f}=this,{rawNode:v}=this.tmNode,p=d("div",Object.assign({class:`${i}-dropdown-option`},r==null?void 0:r(v)),d("div",{class:`${i}-dropdown-option-body ${i}-dropdown-option-body--group`},d("div",{"data-dropdown-option":!0,class:[`${i}-dropdown-option-body__prefix`,t&&`${i}-dropdown-option-body__prefix--show-icon`]},W(v.icon)),d("div",{class:`${i}-dropdown-option-body__label`,"data-dropdown-option":!0},n?n(v):W((e=v.title)!==null&&e!==void 0?e:v[this.labelField])),d("div",{class:[`${i}-dropdown-option-body__suffix`,a&&`${i}-dropdown-option-body__suffix--has-submenu`],"data-dropdown-option":!0})));return f?f({node:p,option:v}):p}});function ne(e,i){return e.type==="submenu"||e.type===void 0&&e[i]!==void 0}function We(e){return e.type==="group"}function ve(e){return e.type==="divider"}function Ve(e){return e.type==="render"}const he=z({name:"DropdownOption",props:{clsPrefix:{type:String,required:!0},tmNode:{type:Object,required:!0},parentKey:{type:[String,Number],default:null},placement:{type:String,default:"right-start"},props:Object,scrollable:Boolean},setup(e){const i=B(V),{hoverKeyRef:a,keyboardKeyRef:t,lastToggledSubmenuKeyRef:r,pendingKeyPathRef:n,activeKeyPathRef:f,animatedRef:v,mergedShowRef:p,renderLabelRef:y,renderIconRef:g,labelFieldRef:S,childrenFieldRef:C,renderOptionRef:N,nodePropsRef:R,menuPropsRef:D}=i,m=B(le,null),I=B(te),_=B(ce),U=b(()=>e.tmNode.rawNode),E=b(()=>{const{value:o}=C;return ne(e.tmNode.rawNode,o)}),X=b(()=>{const{disabled:o}=e.tmNode;return o}),Y=b(()=>{if(!E.value)return!1;const{key:o,disabled:u}=e.tmNode;if(u)return!1;const{value:w}=a,{value:$}=t,{value:ee}=r,{value:A}=n;return w!==null?A.includes(o):$!==null?A.includes(o)&&A[A.length-1]!==o:ee!==null?A.includes(o):!1}),Z=b(()=>t.value===null&&!v.value),J=Ue(Y,300,Z),Q=b(()=>!!(m!=null&&m.enteringSubmenuRef.value)),j=T(!1);H(le,{enteringSubmenuRef:j});function L(){j.value=!0}function q(){j.value=!1}function P(){const{parentKey:o,tmNode:u}=e;u.disabled||p.value&&(r.value=o,t.value=null,a.value=u.key)}function l(){const{tmNode:o}=e;o.disabled||p.value&&a.value!==o.key&&P()}function s(o){if(e.tmNode.disabled||!p.value)return;const{relatedTarget:u}=o;u&&!de({target:u},"dropdownOption")&&!de({target:u},"scrollbarRail")&&(a.value=null)}function c(){const{value:o}=E,{tmNode:u}=e;p.value&&!o&&!u.disabled&&(i.doSelect(u.key,u.rawNode),i.doUpdateShow(!1))}return{labelField:S,renderLabel:y,renderIcon:g,siblingHasIcon:I.showIconRef,siblingHasSubmenu:I.hasSubmenuRef,menuProps:D,popoverBody:_,animated:v,mergedShowSubmenu:b(()=>J.value&&!Q.value),rawNode:U,hasSubmenu:E,pending:G(()=>{const{value:o}=n,{key:u}=e.tmNode;return o.includes(u)}),childActive:G(()=>{const{value:o}=f,{key:u}=e.tmNode,w=o.findIndex($=>u===$);return w===-1?!1:w<o.length-1}),active:G(()=>{const{value:o}=f,{key:u}=e.tmNode,w=o.findIndex($=>u===$);return w===-1?!1:w===o.length-1}),mergedDisabled:X,renderOption:N,nodeProps:R,handleClick:c,handleMouseMove:l,handleMouseEnter:P,handleMouseLeave:s,handleSubmenuBeforeEnter:L,handleSubmenuAfterEnter:q}},render(){var e,i;const{animated:a,rawNode:t,mergedShowSubmenu:r,clsPrefix:n,siblingHasIcon:f,siblingHasSubmenu:v,renderLabel:p,renderIcon:y,renderOption:g,nodeProps:S,props:C,scrollable:N}=this;let R=null;if(r){const _=(e=this.menuProps)===null||e===void 0?void 0:e.call(this,t,t.children);R=d(be,Object.assign({},_,{clsPrefix:n,scrollable:this.scrollable,tmNodes:this.tmNode.children,parentKey:this.tmNode.key}))}const D={class:[`${n}-dropdown-option-body`,this.pending&&`${n}-dropdown-option-body--pending`,this.active&&`${n}-dropdown-option-body--active`,this.childActive&&`${n}-dropdown-option-body--child-active`,this.mergedDisabled&&`${n}-dropdown-option-body--disabled`],onMousemove:this.handleMouseMove,onMouseenter:this.handleMouseEnter,onMouseleave:this.handleMouseLeave,onClick:this.handleClick},m=S==null?void 0:S(t),I=d("div",Object.assign({class:[`${n}-dropdown-option`,m==null?void 0:m.class],"data-dropdown-option":!0},m),d("div",ue(D,C),[d("div",{class:[`${n}-dropdown-option-body__prefix`,f&&`${n}-dropdown-option-body__prefix--show-icon`]},[y?y(t):W(t.icon)]),d("div",{"data-dropdown-option":!0,class:`${n}-dropdown-option-body__label`},p?p(t):W((i=t[this.labelField])!==null&&i!==void 0?i:t.title)),d("div",{"data-dropdown-option":!0,class:[`${n}-dropdown-option-body__suffix`,v&&`${n}-dropdown-option-body__suffix--has-submenu`]},this.hasSubmenu?d(je,null,{default:()=>d(qe,null)}):null)]),this.hasSubmenu?d(Ne,null,{default:()=>[d(Re,null,{default:()=>d("div",{class:`${n}-dropdown-offset-container`},d(Pe,{show:this.mergedShowSubmenu,placement:this.placement,to:N&&this.popoverBody||void 0,teleportDisabled:!N},{default:()=>d("div",{class:`${n}-dropdown-menu-wrapper`},a?d(Ce,{onBeforeEnter:this.handleSubmenuBeforeEnter,onAfterEnter:this.handleSubmenuAfterEnter,name:"fade-in-scale-up-transition",appear:!0},{default:()=>R}):R)}))})]}):null);return g?g({node:I,option:t}):I}}),Xe=z({name:"NDropdownGroup",props:{clsPrefix:{type:String,required:!0},tmNode:{type:Object,required:!0},parentKey:{type:[String,Number],default:null}},render(){const{tmNode:e,parentKey:i,clsPrefix:a}=this,{children:t}=e;return d(Ie,null,d(Ge,{clsPrefix:a,tmNode:e,key:e.key}),t==null?void 0:t.map(r=>{const{rawNode:n}=r;return n.show===!1?null:ve(n)?d(fe,{clsPrefix:a,key:r.key}):r.isGroup?(_e("dropdown","`group` node is not allowed to be put in `group` node."),null):d(he,{clsPrefix:a,tmNode:r,parentKey:i,key:r.key})}))}}),Ye=z({name:"DropdownRenderOption",props:{tmNode:{type:Object,required:!0}},render(){const{rawNode:{render:e,props:i}}=this.tmNode;return d("div",i,[e==null?void 0:e()])}}),be=z({name:"DropdownMenu",props:{scrollable:Boolean,showArrow:Boolean,arrowStyle:[String,Object],clsPrefix:{type:String,required:!0},tmNodes:{type:Array,default:()=>[]},parentKey:{type:[String,Number],default:null}},setup(e){const{renderIconRef:i,childrenFieldRef:a}=B(V);H(te,{showIconRef:b(()=>{const r=i.value;return e.tmNodes.some(n=>{var f;if(n.isGroup)return(f=n.children)===null||f===void 0?void 0:f.some(({rawNode:p})=>r?r(p):p.icon);const{rawNode:v}=n;return r?r(v):v.icon})}),hasSubmenuRef:b(()=>{const{value:r}=a;return e.tmNodes.some(n=>{var f;if(n.isGroup)return(f=n.children)===null||f===void 0?void 0:f.some(({rawNode:p})=>ne(p,r));const{rawNode:v}=n;return ne(v,r)})})});const t=T(null);return H(ze,null),H(De,null),H(ce,t),{bodyRef:t}},render(){const{parentKey:e,clsPrefix:i,scrollable:a}=this,t=this.tmNodes.map(r=>{const{rawNode:n}=r;return n.show===!1?null:Ve(n)?d(Ye,{tmNode:r,key:r.key}):ve(n)?d(fe,{clsPrefix:i,key:r.key}):We(n)?d(Xe,{clsPrefix:i,tmNode:r,parentKey:e,key:r.key}):d(he,{clsPrefix:i,tmNode:r,parentKey:e,key:r.key,props:n.props,scrollable:a})});return d("div",{class:[`${i}-dropdown-menu`,a&&`${i}-dropdown-menu--scrollable`],ref:"bodyRef"},a?d(Oe,{contentClass:`${i}-dropdown-menu__content`},{default:()=>t}):t,this.showArrow?ke({clsPrefix:i,arrowStyle:this.arrowStyle,arrowClass:void 0,arrowWrapperClass:void 0,arrowWrapperStyle:void 0}):null)}}),Ze=x("dropdown-menu",`
 transform-origin: var(--v-transform-origin);
 background-color: var(--n-color);
 border-radius: var(--n-border-radius);
 box-shadow: var(--n-box-shadow);
 position: relative;
 transition:
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
`,[$e(),x("dropdown-option",`
 position: relative;
 `,[M("a",`
 text-decoration: none;
 color: inherit;
 outline: none;
 `,[M("&::before",`
 content: "";
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),x("dropdown-option-body",`
 display: flex;
 cursor: pointer;
 position: relative;
 height: var(--n-option-height);
 line-height: var(--n-option-height);
 font-size: var(--n-font-size);
 color: var(--n-option-text-color);
 transition: color .3s var(--n-bezier);
 `,[M("&::before",`
 content: "";
 position: absolute;
 top: 0;
 bottom: 0;
 left: 4px;
 right: 4px;
 transition: background-color .3s var(--n-bezier);
 border-radius: var(--n-border-radius);
 `),ie("disabled",[k("pending",`
 color: var(--n-option-text-color-hover);
 `,[O("prefix, suffix",`
 color: var(--n-option-text-color-hover);
 `),M("&::before","background-color: var(--n-option-color-hover);")]),k("active",`
 color: var(--n-option-text-color-active);
 `,[O("prefix, suffix",`
 color: var(--n-option-text-color-active);
 `),M("&::before","background-color: var(--n-option-color-active);")]),k("child-active",`
 color: var(--n-option-text-color-child-active);
 `,[O("prefix, suffix",`
 color: var(--n-option-text-color-child-active);
 `)])]),k("disabled",`
 cursor: not-allowed;
 opacity: var(--n-option-opacity-disabled);
 `),k("group",`
 font-size: calc(var(--n-font-size) - 1px);
 color: var(--n-group-header-text-color);
 `,[O("prefix",`
 width: calc(var(--n-option-prefix-width) / 2);
 `,[k("show-icon",`
 width: calc(var(--n-option-icon-prefix-width) / 2);
 `)])]),O("prefix",`
 width: var(--n-option-prefix-width);
 display: flex;
 justify-content: center;
 align-items: center;
 color: var(--n-prefix-color);
 transition: color .3s var(--n-bezier);
 z-index: 1;
 `,[k("show-icon",`
 width: var(--n-option-icon-prefix-width);
 `),x("icon",`
 font-size: var(--n-option-icon-size);
 `)]),O("label",`
 white-space: nowrap;
 flex: 1;
 z-index: 1;
 `),O("suffix",`
 box-sizing: border-box;
 flex-grow: 0;
 flex-shrink: 0;
 display: flex;
 justify-content: flex-end;
 align-items: center;
 min-width: var(--n-option-suffix-width);
 padding: 0 8px;
 transition: color .3s var(--n-bezier);
 color: var(--n-suffix-color);
 z-index: 1;
 `,[k("has-submenu",`
 width: var(--n-option-icon-suffix-width);
 `),x("icon",`
 font-size: var(--n-option-icon-size);
 `)]),x("dropdown-menu","pointer-events: all;")]),x("dropdown-offset-container",`
 pointer-events: none;
 position: absolute;
 left: 0;
 right: 0;
 top: -4px;
 bottom: -4px;
 `)]),x("dropdown-divider",`
 transition: background-color .3s var(--n-bezier);
 background-color: var(--n-divider-color);
 height: 1px;
 margin: 4px 0;
 `),x("dropdown-menu-wrapper",`
 transform-origin: var(--v-transform-origin);
 width: fit-content;
 `),M(">",[x("scrollbar",`
 height: inherit;
 max-height: inherit;
 `)]),ie("scrollable",`
 padding: var(--n-padding);
 `),k("scrollable",[O("content",`
 padding: var(--n-padding);
 `)])]),Je={animated:{type:Boolean,default:!0},keyboard:{type:Boolean,default:!0},size:String,inverted:Boolean,placement:{type:String,default:"bottom"},onSelect:[Function,Array],options:{type:Array,default:()=>[]},menuProps:Function,showArrow:Boolean,renderLabel:Function,renderIcon:Function,renderOption:Function,nodeProps:Function,labelField:{type:String,default:"label"},keyField:{type:String,default:"key"},childrenField:{type:String,default:"children"},value:[String,Number]},Qe=Object.keys(ae),eo=Object.assign(Object.assign(Object.assign({},ae),Je),pe.props),uo=z({name:"Dropdown",inheritAttrs:!1,props:eo,setup(e){const i=T(!1),a=Le(K(e,"show"),i),t=b(()=>{const{keyField:l,childrenField:s}=e;return Ee(e.options,{getKey(c){return c[l]},getDisabled(c){return c.disabled===!0},getIgnored(c){return c.type==="divider"||c.type==="render"},getChildren(c){return c[s]}})}),r=b(()=>t.value.treeNodes),n=T(null),f=T(null),v=T(null),p=b(()=>{var l,s,c;return(c=(s=(l=n.value)!==null&&l!==void 0?l:f.value)!==null&&s!==void 0?s:v.value)!==null&&c!==void 0?c:null}),y=b(()=>t.value.getPath(p.value).keyPath),g=b(()=>t.value.getPath(e.value).keyPath),S=G(()=>e.keyboard&&a.value);Me({keydown:{ArrowUp:{prevent:!0,handler:Z},ArrowRight:{prevent:!0,handler:Y},ArrowDown:{prevent:!0,handler:J},ArrowLeft:{prevent:!0,handler:X},Enter:{prevent:!0,handler:Q},Escape:E}},S);const{mergedClsPrefixRef:C,inlineThemeDisabled:N,mergedComponentPropsRef:R}=Fe(e),D=b(()=>{var l,s;return e.size||((s=(l=R==null?void 0:R.value)===null||l===void 0?void 0:l.Dropdown)===null||s===void 0?void 0:s.size)||"medium"}),m=pe("Dropdown","-dropdown",Ze,Be,e,C);H(V,{labelFieldRef:K(e,"labelField"),childrenFieldRef:K(e,"childrenField"),renderLabelRef:K(e,"renderLabel"),renderIconRef:K(e,"renderIcon"),hoverKeyRef:n,keyboardKeyRef:f,lastToggledSubmenuKeyRef:v,pendingKeyPathRef:y,activeKeyPathRef:g,animatedRef:K(e,"animated"),mergedShowRef:a,nodePropsRef:K(e,"nodeProps"),renderOptionRef:K(e,"renderOption"),menuPropsRef:K(e,"menuProps"),doSelect:I,doUpdateShow:_}),se(a,l=>{!e.animated&&!l&&U()});function I(l,s){const{onSelect:c}=e;c&&oe(c,l,s)}function _(l){const{"onUpdate:show":s,onUpdateShow:c}=e;s&&oe(s,l),c&&oe(c,l),i.value=l}function U(){n.value=null,f.value=null,v.value=null}function E(){_(!1)}function X(){L("left")}function Y(){L("right")}function Z(){L("up")}function J(){L("down")}function Q(){const l=j();l!=null&&l.isLeaf&&a.value&&(I(l.key,l.rawNode),_(!1))}function j(){var l;const{value:s}=t,{value:c}=p;return!s||c===null?null:(l=s.getNode(c))!==null&&l!==void 0?l:null}function L(l){const{value:s}=p,{value:{getFirstAvailableNode:c}}=t;let o=null;if(s===null){const u=c();u!==null&&(o=u.key)}else{const u=j();if(u){let w;switch(l){case"down":w=u.getNext();break;case"up":w=u.getPrev();break;case"right":w=u.getChild();break;case"left":w=u.getParent();break}w&&(o=w.key)}}o!==null&&(n.value=null,f.value=o)}const q=b(()=>{const{inverted:l}=e,s=D.value,{common:{cubicBezierEaseInOut:c},self:o}=m.value,{padding:u,dividerColor:w,borderRadius:$,optionOpacityDisabled:ee,[F("optionIconSuffixWidth",s)]:A,[F("optionSuffixWidth",s)]:we,[F("optionIconPrefixWidth",s)]:me,[F("optionPrefixWidth",s)]:ye,[F("fontSize",s)]:ge,[F("optionHeight",s)]:xe,[F("optionIconSize",s)]:Se}=o,h={"--n-bezier":c,"--n-font-size":ge,"--n-padding":u,"--n-border-radius":$,"--n-option-height":xe,"--n-option-prefix-width":ye,"--n-option-icon-prefix-width":me,"--n-option-suffix-width":we,"--n-option-icon-suffix-width":A,"--n-option-icon-size":Se,"--n-divider-color":w,"--n-option-opacity-disabled":ee};return l?(h["--n-color"]=o.colorInverted,h["--n-option-color-hover"]=o.optionColorHoverInverted,h["--n-option-color-active"]=o.optionColorActiveInverted,h["--n-option-text-color"]=o.optionTextColorInverted,h["--n-option-text-color-hover"]=o.optionTextColorHoverInverted,h["--n-option-text-color-active"]=o.optionTextColorActiveInverted,h["--n-option-text-color-child-active"]=o.optionTextColorChildActiveInverted,h["--n-prefix-color"]=o.prefixColorInverted,h["--n-suffix-color"]=o.suffixColorInverted,h["--n-group-header-text-color"]=o.groupHeaderTextColorInverted):(h["--n-color"]=o.color,h["--n-option-color-hover"]=o.optionColorHover,h["--n-option-color-active"]=o.optionColorActive,h["--n-option-text-color"]=o.optionTextColor,h["--n-option-text-color-hover"]=o.optionTextColorHover,h["--n-option-text-color-active"]=o.optionTextColorActive,h["--n-option-text-color-child-active"]=o.optionTextColorChildActive,h["--n-prefix-color"]=o.prefixColor,h["--n-suffix-color"]=o.suffixColor,h["--n-group-header-text-color"]=o.groupHeaderTextColor),h}),P=N?Te("dropdown",b(()=>`${D.value[0]}${e.inverted?"i":""}`),q,e):void 0;return{mergedClsPrefix:C,mergedTheme:m,mergedSize:D,tmNodes:r,mergedShow:a,handleAfterLeave:()=>{e.animated&&U()},doUpdateShow:_,cssVars:N?void 0:q,themeClass:P==null?void 0:P.themeClass,onRender:P==null?void 0:P.onRender}},render(){const e=(t,r,n,f,v)=>{var p;const{mergedClsPrefix:y,menuProps:g}=this;(p=this.onRender)===null||p===void 0||p.call(this);const S=(g==null?void 0:g(void 0,this.tmNodes.map(N=>N.rawNode)))||{},C={ref:He(r),class:[t,`${y}-dropdown`,`${y}-dropdown--${this.mergedSize}-size`,this.themeClass],clsPrefix:y,tmNodes:this.tmNodes,style:[...n,this.cssVars],showArrow:this.showArrow,arrowStyle:this.arrowStyle,scrollable:this.scrollable,onMouseenter:f,onMouseleave:v};return d(be,ue(this.$attrs,C,S))},{mergedTheme:i}=this,a={show:this.mergedShow,theme:i.peers.Popover,themeOverrides:i.peerOverrides.Popover,internalOnAfterLeave:this.handleAfterLeave,internalRenderBody:e,onUpdateShow:this.doUpdateShow,"onUpdate:show":void 0};return d(Ke,Object.assign({},Ae(this.$props,Qe),a),{trigger:()=>{var t,r;return(r=(t=this.$slots).default)===null||r===void 0?void 0:r.call(t)}})}});export{qe as C,uo as N};
