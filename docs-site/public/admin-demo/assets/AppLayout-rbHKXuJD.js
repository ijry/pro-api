import{ak as E,aS as u,y as c,x,D as P,A as d,cs as te,cG as j,cH as re,v as no,a2 as g,aa as Z,bS as Y,cl as oe,bX as L,bG as lo,bH as io,b4 as ao,aX as V,c4 as so,ae as co,ca as uo,an as vo,a1 as Re,S as He,cA as Oe,d as Ee,H as M,G as Q,aB as mo,bY as X,bj as be,cx as he,b as $e,bi as ce,g as ho,V as fo,cM as ke,a9 as po,bn as go,bm as bo,cu as xo,cr as Co,cI as yo,a6 as W,cN as H,cq as w,af as O,a5 as de,ad as ee,B as ie,a8 as Pe,bZ as zo,bw as So,ck as ue,c3 as Io,cB as wo,c2 as Ro,cC as ko,bI as G}from"./index-BMyk45kF.js";import{N as Ne}from"./text-CmPm_fBo.js";import{N as Po}from"./Tooltip-CP31tRHA.js";import{C as No,N as Le}from"./Dropdown-7lYHFm0t.js";import{f as ve,u as fe}from"./get-D86kArqK.js";import{V as To,a as me}from"./create-BM_mgYTJ.js";import{u as Ao}from"./cssr-E7GQDMMK.js";import{N as Te}from"./Icon-Cx7n14g8.js";import{N as _o}from"./Space-D5OWEM5p.js";import{N as Bo}from"./Tag-1HdK0peN.js";import"./Popover-BQ0-iUw6.js";import"./happens-in-CM8LO42l.js";import"./use-keyboard-B03E-gRL.js";import"./create-ref-setter-C4J8sofl.js";import"./get-slot-Bk_rJcZu.js";const Ho=E({name:"ChevronDownFilled",render(){return u("svg",{viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg"},u("path",{d:"M3.20041 5.73966C3.48226 5.43613 3.95681 5.41856 4.26034 5.70041L8 9.22652L11.7397 5.70041C12.0432 5.41856 12.5177 5.43613 12.7996 5.73966C13.0815 6.0432 13.0639 6.51775 12.7603 6.7996L8.51034 10.7996C8.22258 11.0668 7.77743 11.0668 7.48967 10.7996L3.23966 6.7996C2.93613 6.51775 2.91856 6.0432 3.20041 5.73966Z",fill:"currentColor"}))}}),Oo=c("breadcrumb",`
 white-space: nowrap;
 cursor: default;
 line-height: var(--n-item-line-height);
`,[x("ul",`
 list-style: none;
 padding: 0;
 margin: 0;
 `),x("a",`
 color: inherit;
 text-decoration: inherit;
 `),c("breadcrumb-item",`
 font-size: var(--n-font-size);
 transition: color .3s var(--n-bezier);
 display: inline-flex;
 align-items: center;
 `,[c("icon",`
 font-size: 18px;
 vertical-align: -.2em;
 transition: color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 `),x("&:not(:last-child)",[P("clickable",[d("link",`
 cursor: pointer;
 `,[x("&:hover",`
 background-color: var(--n-item-color-hover);
 `),x("&:active",`
 background-color: var(--n-item-color-pressed); 
 `)])])]),d("link",`
 padding: 4px;
 border-radius: var(--n-item-border-radius);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 position: relative;
 `,[x("&:hover",`
 color: var(--n-item-text-color-hover);
 `,[c("icon",`
 color: var(--n-item-text-color-hover);
 `)]),x("&:active",`
 color: var(--n-item-text-color-pressed);
 `,[c("icon",`
 color: var(--n-item-text-color-pressed);
 `)])]),d("separator",`
 margin: 0 8px;
 color: var(--n-separator-color);
 transition: color .3s var(--n-bezier);
 user-select: none;
 -webkit-user-select: none;
 `),x("&:last-child",[d("link",`
 font-weight: var(--n-font-weight-active);
 cursor: unset;
 color: var(--n-item-text-color-active);
 `,[c("icon",`
 color: var(--n-item-text-color-active);
 `)]),d("separator",`
 display: none;
 `)])])]),Fe=Z("n-breadcrumb"),Eo=Object.assign(Object.assign({},j.props),{separator:{type:String,default:"/"}}),$o=E({name:"Breadcrumb",props:Eo,setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:t}=te(e),a=j("Breadcrumb","-breadcrumb",Oo,no,e,o);Y(Fe,{separatorRef:oe(e,"separator"),mergedClsPrefixRef:o});const i=g(()=>{const{common:{cubicBezierEaseInOut:v},self:{separatorColor:m,itemTextColor:s,itemTextColorHover:h,itemTextColorPressed:_,itemTextColorActive:N,fontSize:f,fontWeightActive:z,itemBorderRadius:S,itemColorHover:C,itemColorPressed:y,itemLineHeight:A}}=a.value;return{"--n-font-size":f,"--n-bezier":v,"--n-item-text-color":s,"--n-item-text-color-hover":h,"--n-item-text-color-pressed":_,"--n-item-text-color-active":N,"--n-separator-color":m,"--n-item-color-hover":C,"--n-item-color-pressed":y,"--n-item-border-radius":S,"--n-font-weight-active":z,"--n-item-line-height":A}}),l=t?re("breadcrumb",void 0,i,e):void 0;return{mergedClsPrefix:o,cssVars:t?void 0:i,themeClass:l==null?void 0:l.themeClass,onRender:l==null?void 0:l.onRender}},render(){var e;return(e=this.onRender)===null||e===void 0||e.call(this),u("nav",{class:[`${this.mergedClsPrefix}-breadcrumb`,this.themeClass],style:this.cssVars,"aria-label":"Breadcrumb"},u("ul",null,this.$slots))}});function Lo(e=ao?window:null){const o=()=>{const{hash:i,host:l,hostname:v,href:m,origin:s,pathname:h,port:_,protocol:N,search:f}=(e==null?void 0:e.location)||{};return{hash:i,host:l,hostname:v,href:m,origin:s,pathname:h,port:_,protocol:N,search:f}},t=L(o()),a=()=>{t.value=o()};return lo(()=>{e&&(e.addEventListener("popstate",a),e.addEventListener("hashchange",a))}),io(()=>{e&&(e.removeEventListener("popstate",a),e.removeEventListener("hashchange",a))}),t}const Fo={separator:String,href:String,clickable:{type:Boolean,default:!0},showSeparator:{type:Boolean,default:!0},onClick:Function},Mo=E({name:"BreadcrumbItem",props:Fo,slots:Object,setup(e,{slots:o}){const t=V(Fe,null);if(!t)return()=>null;const{separatorRef:a,mergedClsPrefixRef:i}=t,l=Lo(),v=g(()=>e.href?"a":"span"),m=g(()=>l.value.href===e.href?"location":null);return()=>{const{value:s}=i;return u("li",{class:[`${s}-breadcrumb-item`,e.clickable&&`${s}-breadcrumb-item--clickable`]},u(v.value,{class:`${s}-breadcrumb-item__link`,"aria-current":m.value,href:e.href,onClick:e.onClick},o),e.showSeparator&&u("span",{class:`${s}-breadcrumb-item__separator`,"aria-hidden":"true"},so(o.separator,()=>{var h;return[(h=e.separator)!==null&&h!==void 0?h:a.value]})))}}});function jo(e){const{baseColor:o,textColor2:t,bodyColor:a,cardColor:i,dividerColor:l,actionColor:v,scrollbarColor:m,scrollbarColorHover:s,invertedColor:h}=e;return{textColor:t,textColorInverted:"#FFF",color:a,colorEmbedded:v,headerColor:i,headerColorInverted:h,footerColor:v,footerColorInverted:h,headerBorderColor:l,headerBorderColorInverted:h,footerBorderColor:l,footerBorderColorInverted:h,siderBorderColor:l,siderBorderColorInverted:h,siderColor:i,siderColorInverted:h,siderToggleButtonBorder:`1px solid ${l}`,siderToggleButtonColor:o,siderToggleButtonIconColor:t,siderToggleButtonIconColorInverted:t,siderToggleBarColor:Re(a,m),siderToggleBarColorHover:Re(a,s),__invertScrollbar:"true"}}const xe=co({name:"Layout",common:vo,peers:{Scrollbar:uo},self:jo}),Me=Z("n-layout-sider"),Ce={type:String,default:"static"},Ko=c("layout",`
 color: var(--n-text-color);
 background-color: var(--n-color);
 box-sizing: border-box;
 position: relative;
 z-index: auto;
 flex: auto;
 overflow: hidden;
 transition:
 box-shadow .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
`,[c("layout-scroll-container",`
 overflow-x: hidden;
 box-sizing: border-box;
 height: 100%;
 `),P("absolute-positioned",`
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),Vo={embedded:Boolean,position:Ce,nativeScrollbar:{type:Boolean,default:!0},scrollbarProps:Object,onScroll:Function,contentClass:String,contentStyle:{type:[String,Object],default:""},hasSider:Boolean,siderPlacement:{type:String,default:"left"}},je=Z("n-layout");function Ke(e){return E({name:e?"LayoutContent":"Layout",props:Object.assign(Object.assign({},j.props),Vo),setup(o){const t=L(null),a=L(null),{mergedClsPrefixRef:i,inlineThemeDisabled:l}=te(o),v=j("Layout","-layout",Ko,xe,o,i);function m(C,y){if(o.nativeScrollbar){const{value:A}=t;A&&(y===void 0?A.scrollTo(C):A.scrollTo(C,y))}else{const{value:A}=a;A&&A.scrollTo(C,y)}}Y(je,o);let s=0,h=0;const _=C=>{var y;const A=C.target;s=A.scrollLeft,h=A.scrollTop,(y=o.onScroll)===null||y===void 0||y.call(o,C)};Oe(()=>{if(o.nativeScrollbar){const C=t.value;C&&(C.scrollTop=h,C.scrollLeft=s)}});const N={display:"flex",flexWrap:"nowrap",width:"100%",flexDirection:"row"},f={scrollTo:m},z=g(()=>{const{common:{cubicBezierEaseInOut:C},self:y}=v.value;return{"--n-bezier":C,"--n-color":o.embedded?y.colorEmbedded:y.color,"--n-text-color":y.textColor}}),S=l?re("layout",g(()=>o.embedded?"e":""),z,o):void 0;return Object.assign({mergedClsPrefix:i,scrollableElRef:t,scrollbarInstRef:a,hasSiderStyle:N,mergedTheme:v,handleNativeElScroll:_,cssVars:l?void 0:z,themeClass:S==null?void 0:S.themeClass,onRender:S==null?void 0:S.onRender},f)},render(){var o;const{mergedClsPrefix:t,hasSider:a}=this;(o=this.onRender)===null||o===void 0||o.call(this);const i=a?this.hasSiderStyle:void 0,l=[this.themeClass,e&&`${t}-layout-content`,`${t}-layout`,`${t}-layout--${this.position}-positioned`];return u("div",{class:l,style:this.cssVars},this.nativeScrollbar?u("div",{ref:"scrollableElRef",class:[`${t}-layout-scroll-container`,this.contentClass],style:[this.contentStyle,i],onScroll:this.handleNativeElScroll},this.$slots):u(He,Object.assign({},this.scrollbarProps,{onScroll:this.onScroll,ref:"scrollbarInstRef",theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,contentClass:this.contentClass,contentStyle:[this.contentStyle,i]}),this.$slots))}})}const Ae=Ke(!1),Do=Ke(!0),Uo=c("layout-header",`
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 box-sizing: border-box;
 width: 100%;
 background-color: var(--n-color);
 color: var(--n-text-color);
`,[P("absolute-positioned",`
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 `),P("bordered",`
 border-bottom: solid 1px var(--n-border-color);
 `)]),Go={position:Ce,inverted:Boolean,bordered:{type:Boolean,default:!1}},qo=E({name:"LayoutHeader",props:Object.assign(Object.assign({},j.props),Go),setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:t}=te(e),a=j("Layout","-layout-header",Uo,xe,e,o),i=g(()=>{const{common:{cubicBezierEaseInOut:v},self:m}=a.value,s={"--n-bezier":v};return e.inverted?(s["--n-color"]=m.headerColorInverted,s["--n-text-color"]=m.textColorInverted,s["--n-border-color"]=m.headerBorderColorInverted):(s["--n-color"]=m.headerColor,s["--n-text-color"]=m.textColor,s["--n-border-color"]=m.headerBorderColor),s}),l=t?re("layout-header",g(()=>e.inverted?"a":"b"),i,e):void 0;return{mergedClsPrefix:o,cssVars:t?void 0:i,themeClass:l==null?void 0:l.themeClass,onRender:l==null?void 0:l.onRender}},render(){var e;const{mergedClsPrefix:o}=this;return(e=this.onRender)===null||e===void 0||e.call(this),u("div",{class:[`${o}-layout-header`,this.themeClass,this.position&&`${o}-layout-header--${this.position}-positioned`,this.bordered&&`${o}-layout-header--bordered`],style:this.cssVars},this.$slots)}}),Yo=c("layout-sider",`
 flex-shrink: 0;
 box-sizing: border-box;
 position: relative;
 z-index: 1;
 color: var(--n-text-color);
 transition:
 color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 min-width .3s var(--n-bezier),
 max-width .3s var(--n-bezier),
 transform .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 background-color: var(--n-color);
 display: flex;
 justify-content: flex-end;
`,[P("bordered",[d("border",`
 content: "";
 position: absolute;
 top: 0;
 bottom: 0;
 width: 1px;
 background-color: var(--n-border-color);
 transition: background-color .3s var(--n-bezier);
 `)]),d("left-placement",[P("bordered",[d("border",`
 right: 0;
 `)])]),P("right-placement",`
 justify-content: flex-start;
 `,[P("bordered",[d("border",`
 left: 0;
 `)]),P("collapsed",[c("layout-toggle-button",[c("base-icon",`
 transform: rotate(180deg);
 `)]),c("layout-toggle-bar",[x("&:hover",[d("top",{transform:"rotate(-12deg) scale(1.15) translateY(-2px)"}),d("bottom",{transform:"rotate(12deg) scale(1.15) translateY(2px)"})])])]),c("layout-toggle-button",`
 left: 0;
 transform: translateX(-50%) translateY(-50%);
 `,[c("base-icon",`
 transform: rotate(0);
 `)]),c("layout-toggle-bar",`
 left: -28px;
 transform: rotate(180deg);
 `,[x("&:hover",[d("top",{transform:"rotate(12deg) scale(1.15) translateY(-2px)"}),d("bottom",{transform:"rotate(-12deg) scale(1.15) translateY(2px)"})])])]),P("collapsed",[c("layout-toggle-bar",[x("&:hover",[d("top",{transform:"rotate(-12deg) scale(1.15) translateY(-2px)"}),d("bottom",{transform:"rotate(12deg) scale(1.15) translateY(2px)"})])]),c("layout-toggle-button",[c("base-icon",`
 transform: rotate(0);
 `)])]),c("layout-toggle-button",`
 transition:
 color .3s var(--n-bezier),
 right .3s var(--n-bezier),
 left .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 cursor: pointer;
 width: 24px;
 height: 24px;
 position: absolute;
 top: 50%;
 right: 0;
 border-radius: 50%;
 display: flex;
 align-items: center;
 justify-content: center;
 font-size: 18px;
 color: var(--n-toggle-button-icon-color);
 border: var(--n-toggle-button-border);
 background-color: var(--n-toggle-button-color);
 box-shadow: 0 2px 4px 0px rgba(0, 0, 0, .06);
 transform: translateX(50%) translateY(-50%);
 z-index: 1;
 `,[c("base-icon",`
 transition: transform .3s var(--n-bezier);
 transform: rotate(180deg);
 `)]),c("layout-toggle-bar",`
 cursor: pointer;
 height: 72px;
 width: 32px;
 position: absolute;
 top: calc(50% - 36px);
 right: -28px;
 `,[d("top, bottom",`
 position: absolute;
 width: 4px;
 border-radius: 2px;
 height: 38px;
 left: 14px;
 transition: 
 background-color .3s var(--n-bezier),
 transform .3s var(--n-bezier);
 `),d("bottom",`
 position: absolute;
 top: 34px;
 `),x("&:hover",[d("top",{transform:"rotate(12deg) scale(1.15) translateY(-2px)"}),d("bottom",{transform:"rotate(-12deg) scale(1.15) translateY(2px)"})]),d("top, bottom",{backgroundColor:"var(--n-toggle-bar-color)"}),x("&:hover",[d("top, bottom",{backgroundColor:"var(--n-toggle-bar-color-hover)"})])]),d("border",`
 position: absolute;
 top: 0;
 right: 0;
 bottom: 0;
 width: 1px;
 transition: background-color .3s var(--n-bezier);
 `),c("layout-sider-scroll-container",`
 flex-grow: 1;
 flex-shrink: 0;
 box-sizing: border-box;
 height: 100%;
 opacity: 0;
 transition: opacity .3s var(--n-bezier);
 max-width: 100%;
 `),P("show-content",[c("layout-sider-scroll-container",{opacity:1})]),P("absolute-positioned",`
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 `)]),Wo=E({props:{clsPrefix:{type:String,required:!0},onClick:Function},render(){const{clsPrefix:e}=this;return u("div",{onClick:this.onClick,class:`${e}-layout-toggle-bar`},u("div",{class:`${e}-layout-toggle-bar__top`}),u("div",{class:`${e}-layout-toggle-bar__bottom`}))}}),Xo=E({name:"LayoutToggleButton",props:{clsPrefix:{type:String,required:!0},onClick:Function},render(){const{clsPrefix:e}=this;return u("div",{class:`${e}-layout-toggle-button`,onClick:this.onClick},u(Ee,{clsPrefix:e},{default:()=>u(No,null)}))}}),Zo={position:Ce,bordered:Boolean,collapsedWidth:{type:Number,default:48},width:{type:[Number,String],default:272},contentClass:String,contentStyle:{type:[String,Object],default:""},collapseMode:{type:String,default:"transform"},collapsed:{type:Boolean,default:void 0},defaultCollapsed:Boolean,showCollapsedContent:{type:Boolean,default:!0},showTrigger:{type:[Boolean,String],default:!1},nativeScrollbar:{type:Boolean,default:!0},inverted:Boolean,scrollbarProps:Object,triggerClass:String,triggerStyle:[String,Object],collapsedTriggerClass:String,collapsedTriggerStyle:[String,Object],"onUpdate:collapsed":[Function,Array],onUpdateCollapsed:[Function,Array],onAfterEnter:Function,onAfterLeave:Function,onExpand:[Function,Array],onCollapse:[Function,Array],onScroll:Function},Jo=E({name:"LayoutSider",props:Object.assign(Object.assign({},j.props),Zo),setup(e){const o=V(je),t=L(null),a=L(null),i=L(e.defaultCollapsed),l=fe(oe(e,"collapsed"),i),v=g(()=>ve(l.value?e.collapsedWidth:e.width)),m=g(()=>e.collapseMode!=="transform"?{}:{minWidth:ve(e.width)}),s=g(()=>o?o.siderPlacement:"left");function h(B,R){if(e.nativeScrollbar){const{value:k}=t;k&&(R===void 0?k.scrollTo(B):k.scrollTo(B,R))}else{const{value:k}=a;k&&k.scrollTo(B,R)}}function _(){const{"onUpdate:collapsed":B,onUpdateCollapsed:R,onExpand:k,onCollapse:U}=e,{value:K}=l;R&&M(R,!K),B&&M(B,!K),i.value=!K,K?k&&M(k):U&&M(U)}let N=0,f=0;const z=B=>{var R;const k=B.target;N=k.scrollLeft,f=k.scrollTop,(R=e.onScroll)===null||R===void 0||R.call(e,B)};Oe(()=>{if(e.nativeScrollbar){const B=t.value;B&&(B.scrollTop=f,B.scrollLeft=N)}}),Y(Me,{collapsedRef:l,collapseModeRef:oe(e,"collapseMode")});const{mergedClsPrefixRef:S,inlineThemeDisabled:C}=te(e),y=j("Layout","-layout-sider",Yo,xe,e,S);function A(B){var R,k;B.propertyName==="max-width"&&(l.value?(R=e.onAfterLeave)===null||R===void 0||R.call(e):(k=e.onAfterEnter)===null||k===void 0||k.call(e))}const J={scrollTo:h},D=g(()=>{const{common:{cubicBezierEaseInOut:B},self:R}=y.value,{siderToggleButtonColor:k,siderToggleButtonBorder:U,siderToggleBarColor:K,siderToggleBarColorHover:se}=R,$={"--n-bezier":B,"--n-toggle-button-color":k,"--n-toggle-button-border":U,"--n-toggle-bar-color":K,"--n-toggle-bar-color-hover":se};return e.inverted?($["--n-color"]=R.siderColorInverted,$["--n-text-color"]=R.textColorInverted,$["--n-border-color"]=R.siderBorderColorInverted,$["--n-toggle-button-icon-color"]=R.siderToggleButtonIconColorInverted,$.__invertScrollbar=R.__invertScrollbar):($["--n-color"]=R.siderColor,$["--n-text-color"]=R.textColor,$["--n-border-color"]=R.siderBorderColor,$["--n-toggle-button-icon-color"]=R.siderToggleButtonIconColor),$}),F=C?re("layout-sider",g(()=>e.inverted?"a":"b"),D,e):void 0;return Object.assign({scrollableElRef:t,scrollbarInstRef:a,mergedClsPrefix:S,mergedTheme:y,styleMaxWidth:v,mergedCollapsed:l,scrollContainerStyle:m,siderPlacement:s,handleNativeElScroll:z,handleTransitionend:A,handleTriggerClick:_,inlineThemeDisabled:C,cssVars:D,themeClass:F==null?void 0:F.themeClass,onRender:F==null?void 0:F.onRender},J)},render(){var e;const{mergedClsPrefix:o,mergedCollapsed:t,showTrigger:a}=this;return(e=this.onRender)===null||e===void 0||e.call(this),u("aside",{class:[`${o}-layout-sider`,this.themeClass,`${o}-layout-sider--${this.position}-positioned`,`${o}-layout-sider--${this.siderPlacement}-placement`,this.bordered&&`${o}-layout-sider--bordered`,t&&`${o}-layout-sider--collapsed`,(!t||this.showCollapsedContent)&&`${o}-layout-sider--show-content`],onTransitionend:this.handleTransitionend,style:[this.inlineThemeDisabled?void 0:this.cssVars,{maxWidth:this.styleMaxWidth,width:ve(this.width)}]},this.nativeScrollbar?u("div",{class:[`${o}-layout-sider-scroll-container`,this.contentClass],onScroll:this.handleNativeElScroll,style:[this.scrollContainerStyle,{overflow:"auto"},this.contentStyle],ref:"scrollableElRef"},this.$slots):u(He,Object.assign({},this.scrollbarProps,{onScroll:this.onScroll,ref:"scrollbarInstRef",style:this.scrollContainerStyle,contentStyle:this.contentStyle,contentClass:this.contentClass,theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,builtinThemeOverrides:this.inverted&&this.cssVars.__invertScrollbar==="true"?{colorHover:"rgba(255, 255, 255, .4)",color:"rgba(255, 255, 255, .3)"}:void 0}),this.$slots),a?a==="bar"?u(Wo,{clsPrefix:o,class:t?this.collapsedTriggerClass:this.triggerClass,style:t?this.collapsedTriggerStyle:this.triggerStyle,onClick:this.handleTriggerClick}):u(Xo,{clsPrefix:o,class:t?this.collapsedTriggerClass:this.triggerClass,style:t?this.collapsedTriggerStyle:this.triggerStyle,onClick:this.handleTriggerClick}):null,this.bordered?u("div",{class:`${o}-layout-sider__border`}):null)}}),ne=Z("n-menu"),Ve=Z("n-submenu"),ye=Z("n-menu-item-group"),_e=[x("&::before","background-color: var(--n-item-color-hover);"),d("arrow",`
 color: var(--n-arrow-color-hover);
 `),d("icon",`
 color: var(--n-item-icon-color-hover);
 `),c("menu-item-content-header",`
 color: var(--n-item-text-color-hover);
 `,[x("a",`
 color: var(--n-item-text-color-hover);
 `),d("extra",`
 color: var(--n-item-text-color-hover);
 `)])],Be=[d("icon",`
 color: var(--n-item-icon-color-hover-horizontal);
 `),c("menu-item-content-header",`
 color: var(--n-item-text-color-hover-horizontal);
 `,[x("a",`
 color: var(--n-item-text-color-hover-horizontal);
 `),d("extra",`
 color: var(--n-item-text-color-hover-horizontal);
 `)])],Qo=x([c("menu",`
 background-color: var(--n-color);
 color: var(--n-item-text-color);
 overflow: hidden;
 transition: background-color .3s var(--n-bezier);
 box-sizing: border-box;
 font-size: var(--n-font-size);
 padding-bottom: 6px;
 `,[P("horizontal",`
 max-width: 100%;
 width: 100%;
 display: flex;
 overflow: hidden;
 padding-bottom: 0;
 `,[c("submenu","margin: 0;"),c("menu-item","margin: 0;"),c("menu-item-content",`
 padding: 0 20px;
 border-bottom: 2px solid #0000;
 `,[x("&::before","display: none;"),P("selected","border-bottom: 2px solid var(--n-border-color-horizontal)")]),c("menu-item-content",[P("selected",[d("icon","color: var(--n-item-icon-color-active-horizontal);"),c("menu-item-content-header",`
 color: var(--n-item-text-color-active-horizontal);
 `,[x("a","color: var(--n-item-text-color-active-horizontal);"),d("extra","color: var(--n-item-text-color-active-horizontal);")])]),P("child-active",`
 border-bottom: 2px solid var(--n-border-color-horizontal);
 `,[c("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-horizontal);
 `,[x("a",`
 color: var(--n-item-text-color-child-active-horizontal);
 `),d("extra",`
 color: var(--n-item-text-color-child-active-horizontal);
 `)]),d("icon",`
 color: var(--n-item-icon-color-child-active-horizontal);
 `)]),Q("disabled",[Q("selected, child-active",[x("&:focus-within",Be)]),P("selected",[q(null,[d("icon","color: var(--n-item-icon-color-active-hover-horizontal);"),c("menu-item-content-header",`
 color: var(--n-item-text-color-active-hover-horizontal);
 `,[x("a","color: var(--n-item-text-color-active-hover-horizontal);"),d("extra","color: var(--n-item-text-color-active-hover-horizontal);")])])]),P("child-active",[q(null,[d("icon","color: var(--n-item-icon-color-child-active-hover-horizontal);"),c("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-hover-horizontal);
 `,[x("a","color: var(--n-item-text-color-child-active-hover-horizontal);"),d("extra","color: var(--n-item-text-color-child-active-hover-horizontal);")])])]),q("border-bottom: 2px solid var(--n-border-color-horizontal);",Be)]),c("menu-item-content-header",[x("a","color: var(--n-item-text-color-horizontal);")])])]),Q("responsive",[c("menu-item-content-header",`
 overflow: hidden;
 text-overflow: ellipsis;
 `)]),P("collapsed",[c("menu-item-content",[P("selected",[x("&::before",`
 background-color: var(--n-item-color-active-collapsed) !important;
 `)]),c("menu-item-content-header","opacity: 0;"),d("arrow","opacity: 0;"),d("icon","color: var(--n-item-icon-color-collapsed);")])]),c("menu-item",`
 height: var(--n-item-height);
 margin-top: 6px;
 position: relative;
 `),c("menu-item-content",`
 box-sizing: border-box;
 line-height: 1.75;
 height: 100%;
 display: grid;
 grid-template-areas: "icon content arrow";
 grid-template-columns: auto 1fr auto;
 align-items: center;
 cursor: pointer;
 position: relative;
 padding-right: 18px;
 transition:
 background-color .3s var(--n-bezier),
 padding-left .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `,[x("> *","z-index: 1;"),x("&::before",`
 z-index: auto;
 content: "";
 background-color: #0000;
 position: absolute;
 left: 8px;
 right: 8px;
 top: 0;
 bottom: 0;
 pointer-events: none;
 border-radius: var(--n-border-radius);
 transition: background-color .3s var(--n-bezier);
 `),P("disabled",`
 opacity: .45;
 cursor: not-allowed;
 `),P("collapsed",[d("arrow","transform: rotate(0);")]),P("selected",[x("&::before","background-color: var(--n-item-color-active);"),d("arrow","color: var(--n-arrow-color-active);"),d("icon","color: var(--n-item-icon-color-active);"),c("menu-item-content-header",`
 color: var(--n-item-text-color-active);
 `,[x("a","color: var(--n-item-text-color-active);"),d("extra","color: var(--n-item-text-color-active);")])]),P("child-active",[c("menu-item-content-header",`
 color: var(--n-item-text-color-child-active);
 `,[x("a",`
 color: var(--n-item-text-color-child-active);
 `),d("extra",`
 color: var(--n-item-text-color-child-active);
 `)]),d("arrow",`
 color: var(--n-arrow-color-child-active);
 `),d("icon",`
 color: var(--n-item-icon-color-child-active);
 `)]),Q("disabled",[Q("selected, child-active",[x("&:focus-within",_e)]),P("selected",[q(null,[d("arrow","color: var(--n-arrow-color-active-hover);"),d("icon","color: var(--n-item-icon-color-active-hover);"),c("menu-item-content-header",`
 color: var(--n-item-text-color-active-hover);
 `,[x("a","color: var(--n-item-text-color-active-hover);"),d("extra","color: var(--n-item-text-color-active-hover);")])])]),P("child-active",[q(null,[d("arrow","color: var(--n-arrow-color-child-active-hover);"),d("icon","color: var(--n-item-icon-color-child-active-hover);"),c("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-hover);
 `,[x("a","color: var(--n-item-text-color-child-active-hover);"),d("extra","color: var(--n-item-text-color-child-active-hover);")])])]),P("selected",[q(null,[x("&::before","background-color: var(--n-item-color-active-hover);")])]),q(null,_e)]),d("icon",`
 grid-area: icon;
 color: var(--n-item-icon-color);
 transition:
 color .3s var(--n-bezier),
 font-size .3s var(--n-bezier),
 margin-right .3s var(--n-bezier);
 box-sizing: content-box;
 display: inline-flex;
 align-items: center;
 justify-content: center;
 `),d("arrow",`
 grid-area: arrow;
 font-size: 16px;
 color: var(--n-arrow-color);
 transform: rotate(180deg);
 opacity: 1;
 transition:
 color .3s var(--n-bezier),
 transform 0.2s var(--n-bezier),
 opacity 0.2s var(--n-bezier);
 `),c("menu-item-content-header",`
 grid-area: content;
 transition:
 color .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 opacity: 1;
 white-space: nowrap;
 color: var(--n-item-text-color);
 `,[x("a",`
 outline: none;
 text-decoration: none;
 transition: color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 `,[x("&::before",`
 content: "";
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),d("extra",`
 font-size: .93em;
 color: var(--n-group-text-color);
 transition: color .3s var(--n-bezier);
 `)])]),c("submenu",`
 cursor: pointer;
 position: relative;
 margin-top: 6px;
 `,[c("menu-item-content",`
 height: var(--n-item-height);
 `),c("submenu-children",`
 overflow: hidden;
 padding: 0;
 `,[mo({duration:".2s"})])]),c("menu-item-group",[c("menu-item-group-title",`
 margin-top: 6px;
 color: var(--n-group-text-color);
 cursor: default;
 font-size: .93em;
 height: 36px;
 display: flex;
 align-items: center;
 transition:
 padding-left .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `)])]),c("menu-tooltip",[x("a",`
 color: inherit;
 text-decoration: none;
 `)]),c("menu-divider",`
 transition: background-color .3s var(--n-bezier);
 background-color: var(--n-divider-color);
 height: 1px;
 margin: 6px 18px;
 `)]);function q(e,o){return[P("hover",e,o),x("&:hover",e,o)]}const De=E({name:"MenuOptionContent",props:{collapsed:Boolean,disabled:Boolean,title:[String,Function],icon:Function,extra:[String,Function],showArrow:Boolean,childActive:Boolean,hover:Boolean,paddingLeft:Number,selected:Boolean,maxIconSize:{type:Number,required:!0},activeIconSize:{type:Number,required:!0},iconMarginRight:{type:Number,required:!0},clsPrefix:{type:String,required:!0},onClick:Function,tmNode:{type:Object,required:!0},isEllipsisPlaceholder:Boolean},setup(e){const{props:o}=V(ne);return{menuProps:o,style:g(()=>{const{paddingLeft:t}=e;return{paddingLeft:t&&`${t}px`}}),iconStyle:g(()=>{const{maxIconSize:t,activeIconSize:a,iconMarginRight:i}=e;return{width:`${t}px`,height:`${t}px`,fontSize:`${a}px`,marginRight:`${i}px`}})}},render(){const{clsPrefix:e,tmNode:o,menuProps:{renderIcon:t,renderLabel:a,renderExtra:i,expandIcon:l}}=this,v=t?t(o.rawNode):X(this.icon);return u("div",{onClick:m=>{var s;(s=this.onClick)===null||s===void 0||s.call(this,m)},role:"none",class:[`${e}-menu-item-content`,{[`${e}-menu-item-content--selected`]:this.selected,[`${e}-menu-item-content--collapsed`]:this.collapsed,[`${e}-menu-item-content--child-active`]:this.childActive,[`${e}-menu-item-content--disabled`]:this.disabled,[`${e}-menu-item-content--hover`]:this.hover}],style:this.style},v&&u("div",{class:`${e}-menu-item-content__icon`,style:this.iconStyle,role:"none"},[v]),u("div",{class:`${e}-menu-item-content-header`,role:"none"},this.isEllipsisPlaceholder?this.title:a?a(o.rawNode):X(this.title),this.extra||i?u("span",{class:`${e}-menu-item-content-header__extra`}," ",i?i(o.rawNode):X(this.extra)):null),this.showArrow?u(Ee,{ariaHidden:!0,class:`${e}-menu-item-content__arrow`,clsPrefix:e},{default:()=>l?l(o.rawNode):u(Ho,null)}):null)}}),ae=8;function ze(e){const o=V(ne),{props:t,mergedCollapsedRef:a}=o,i=V(Ve,null),l=V(ye,null),v=g(()=>t.mode==="horizontal"),m=g(()=>v.value?t.dropdownPlacement:"tmNodes"in e?"right-start":"right"),s=g(()=>{var f;return Math.max((f=t.collapsedIconSize)!==null&&f!==void 0?f:t.iconSize,t.iconSize)}),h=g(()=>{var f;return!v.value&&e.root&&a.value&&(f=t.collapsedIconSize)!==null&&f!==void 0?f:t.iconSize}),_=g(()=>{if(v.value)return;const{collapsedWidth:f,indent:z,rootIndent:S}=t,{root:C,isGroup:y}=e,A=S===void 0?z:S;return C?a.value?f/2-s.value/2:A:l&&typeof l.paddingLeftRef.value=="number"?z/2+l.paddingLeftRef.value:i&&typeof i.paddingLeftRef.value=="number"?(y?z/2:z)+i.paddingLeftRef.value:0}),N=g(()=>{const{collapsedWidth:f,indent:z,rootIndent:S}=t,{value:C}=s,{root:y}=e;return v.value||!y||!a.value?ae:(S===void 0?z:S)+C+ae-(f+C)/2});return{dropdownPlacement:m,activeIconSize:h,maxIconSize:s,paddingLeft:_,iconMarginRight:N,NMenu:o,NSubmenu:i,NMenuOptionGroup:l}}const Se={internalKey:{type:[String,Number],required:!0},root:Boolean,isGroup:Boolean,level:{type:Number,required:!0},title:[String,Function],extra:[String,Function]},et=E({name:"MenuDivider",setup(){const e=V(ne),{mergedClsPrefixRef:o,isHorizontalRef:t}=e;return()=>t.value?null:u("div",{class:`${o.value}-menu-divider`})}}),Ue=Object.assign(Object.assign({},Se),{tmNode:{type:Object,required:!0},disabled:Boolean,icon:Function,onClick:Function}),ot=be(Ue),tt=E({name:"MenuOption",props:Ue,setup(e){const o=ze(e),{NSubmenu:t,NMenu:a,NMenuOptionGroup:i}=o,{props:l,mergedClsPrefixRef:v,mergedCollapsedRef:m}=a,s=t?t.mergedDisabledRef:i?i.mergedDisabledRef:{value:!1},h=g(()=>s.value||e.disabled);function _(f){const{onClick:z}=e;z&&z(f)}function N(f){h.value||(a.doSelect(e.internalKey,e.tmNode.rawNode),_(f))}return{mergedClsPrefix:v,dropdownPlacement:o.dropdownPlacement,paddingLeft:o.paddingLeft,iconMarginRight:o.iconMarginRight,maxIconSize:o.maxIconSize,activeIconSize:o.activeIconSize,mergedTheme:a.mergedThemeRef,menuProps:l,dropdownEnabled:he(()=>e.root&&m.value&&l.mode!=="horizontal"&&!h.value),selected:he(()=>a.mergedValueRef.value===e.internalKey),mergedDisabled:h,handleClick:N}},render(){const{mergedClsPrefix:e,mergedTheme:o,tmNode:t,menuProps:{renderLabel:a,nodeProps:i}}=this,l=i==null?void 0:i(t.rawNode);return u("div",Object.assign({},l,{role:"menuitem",class:[`${e}-menu-item`,l==null?void 0:l.class]}),u(Po,{theme:o.peers.Tooltip,themeOverrides:o.peerOverrides.Tooltip,trigger:"hover",placement:this.dropdownPlacement,disabled:!this.dropdownEnabled||this.title===void 0,internalExtraClass:["menu-tooltip"]},{default:()=>a?a(t.rawNode):X(this.title),trigger:()=>u(De,{tmNode:t,clsPrefix:e,paddingLeft:this.paddingLeft,iconMarginRight:this.iconMarginRight,maxIconSize:this.maxIconSize,activeIconSize:this.activeIconSize,selected:this.selected,title:this.title,extra:this.extra,disabled:this.mergedDisabled,icon:this.icon,onClick:this.handleClick})}))}}),Ge=Object.assign(Object.assign({},Se),{tmNode:{type:Object,required:!0},tmNodes:{type:Array,required:!0}}),rt=be(Ge),nt=E({name:"MenuOptionGroup",props:Ge,setup(e){const o=ze(e),{NSubmenu:t}=o,a=g(()=>t!=null&&t.mergedDisabledRef.value?!0:e.tmNode.disabled);Y(ye,{paddingLeftRef:o.paddingLeft,mergedDisabledRef:a});const{mergedClsPrefixRef:i,props:l}=V(ne);return function(){const{value:v}=i,m=o.paddingLeft.value,{nodeProps:s}=l,h=s==null?void 0:s(e.tmNode.rawNode);return u("div",{class:`${v}-menu-item-group`,role:"group"},u("div",Object.assign({},h,{class:[`${v}-menu-item-group-title`,h==null?void 0:h.class],style:[(h==null?void 0:h.style)||"",m!==void 0?`padding-left: ${m}px;`:""]}),X(e.title),e.extra?u($e,null," ",X(e.extra)):null),u("div",null,e.tmNodes.map(_=>Ie(_,l))))}}});function pe(e){return e.type==="divider"||e.type==="render"}function lt(e){return e.type==="divider"}function Ie(e,o){const{rawNode:t}=e,{show:a}=t;if(a===!1)return null;if(pe(t))return lt(t)?u(et,Object.assign({key:e.key},t.props)):null;const{labelField:i}=o,{key:l,level:v,isGroup:m}=e,s=Object.assign(Object.assign({},t),{title:t.title||t[i],extra:t.titleExtra||t.extra,key:l,internalKey:l,level:v,root:v===0,isGroup:m});return e.children?e.isGroup?u(nt,ce(s,rt,{tmNode:e,tmNodes:e.children,key:l})):u(ge,ce(s,it,{key:l,rawNodes:t[o.childrenField],tmNodes:e.children,tmNode:e})):u(tt,ce(s,ot,{key:l,tmNode:e}))}const qe=Object.assign(Object.assign({},Se),{rawNodes:{type:Array,default:()=>[]},tmNodes:{type:Array,default:()=>[]},tmNode:{type:Object,required:!0},disabled:Boolean,icon:Function,onClick:Function,domId:String,virtualChildActive:{type:Boolean,default:void 0},isEllipsisPlaceholder:Boolean}),it=be(qe),ge=E({name:"Submenu",props:qe,setup(e){const o=ze(e),{NMenu:t,NSubmenu:a}=o,{props:i,mergedCollapsedRef:l,mergedThemeRef:v}=t,m=g(()=>{const{disabled:f}=e;return a!=null&&a.mergedDisabledRef.value||i.disabled?!0:f}),s=L(!1);Y(Ve,{paddingLeftRef:o.paddingLeft,mergedDisabledRef:m}),Y(ye,null);function h(){const{onClick:f}=e;f&&f()}function _(){m.value||(l.value||t.toggleExpand(e.internalKey),h())}function N(f){s.value=f}return{menuProps:i,mergedTheme:v,doSelect:t.doSelect,inverted:t.invertedRef,isHorizontal:t.isHorizontalRef,mergedClsPrefix:t.mergedClsPrefixRef,maxIconSize:o.maxIconSize,activeIconSize:o.activeIconSize,iconMarginRight:o.iconMarginRight,dropdownPlacement:o.dropdownPlacement,dropdownShow:s,paddingLeft:o.paddingLeft,mergedDisabled:m,mergedValue:t.mergedValueRef,childActive:he(()=>{var f;return(f=e.virtualChildActive)!==null&&f!==void 0?f:t.activePathRef.value.includes(e.internalKey)}),collapsed:g(()=>i.mode==="horizontal"?!1:l.value?!0:!t.mergedExpandedKeysRef.value.includes(e.internalKey)),dropdownEnabled:g(()=>!m.value&&(i.mode==="horizontal"||l.value)),handlePopoverShowChange:N,handleClick:_}},render(){var e;const{mergedClsPrefix:o,menuProps:{renderIcon:t,renderLabel:a}}=this,i=()=>{const{isHorizontal:v,paddingLeft:m,collapsed:s,mergedDisabled:h,maxIconSize:_,activeIconSize:N,title:f,childActive:z,icon:S,handleClick:C,menuProps:{nodeProps:y},dropdownShow:A,iconMarginRight:J,tmNode:D,mergedClsPrefix:F,isEllipsisPlaceholder:B,extra:R}=this,k=y==null?void 0:y(D.rawNode);return u("div",Object.assign({},k,{class:[`${F}-menu-item`,k==null?void 0:k.class],role:"menuitem"}),u(De,{tmNode:D,paddingLeft:m,collapsed:s,disabled:h,iconMarginRight:J,maxIconSize:_,activeIconSize:N,title:f,extra:R,showArrow:!v,childActive:z,clsPrefix:F,icon:S,hover:A,onClick:C,isEllipsisPlaceholder:B}))},l=()=>u(ho,null,{default:()=>{const{tmNodes:v,collapsed:m}=this;return m?null:u("div",{class:`${o}-submenu-children`,role:"menu"},v.map(s=>Ie(s,this.menuProps)))}});return this.root?u(Le,Object.assign({size:"large",trigger:"hover"},(e=this.menuProps)===null||e===void 0?void 0:e.dropdownProps,{themeOverrides:this.mergedTheme.peerOverrides.Dropdown,theme:this.mergedTheme.peers.Dropdown,builtinThemeOverrides:{fontSizeLarge:"14px",optionIconSizeLarge:"18px"},value:this.mergedValue,disabled:!this.dropdownEnabled,placement:this.dropdownPlacement,keyField:this.menuProps.keyField,labelField:this.menuProps.labelField,childrenField:this.menuProps.childrenField,onUpdateShow:this.handlePopoverShowChange,options:this.rawNodes,onSelect:this.doSelect,inverted:this.inverted,renderIcon:t,renderLabel:a}),{default:()=>u("div",{class:`${o}-submenu`,role:"menu","aria-expanded":!this.collapsed,id:this.domId},i(),this.isHorizontal?null:l())}):u("div",{class:`${o}-submenu`,role:"menu","aria-expanded":!this.collapsed,id:this.domId},i(),l())}}),at=Object.assign(Object.assign({},j.props),{options:{type:Array,default:()=>[]},collapsed:{type:Boolean,default:void 0},collapsedWidth:{type:Number,default:48},iconSize:{type:Number,default:20},collapsedIconSize:{type:Number,default:24},rootIndent:Number,indent:{type:Number,default:32},labelField:{type:String,default:"label"},keyField:{type:String,default:"key"},childrenField:{type:String,default:"children"},disabledField:{type:String,default:"disabled"},defaultExpandAll:Boolean,defaultExpandedKeys:Array,expandedKeys:Array,value:[String,Number],defaultValue:{type:[String,Number],default:null},mode:{type:String,default:"vertical"},watchProps:{type:Array,default:void 0},disabled:Boolean,show:{type:Boolean,default:!0},inverted:Boolean,"onUpdate:expandedKeys":[Function,Array],onUpdateExpandedKeys:[Function,Array],onUpdateValue:[Function,Array],"onUpdate:value":[Function,Array],expandIcon:Function,renderIcon:Function,renderLabel:Function,renderExtra:Function,dropdownProps:Object,accordion:Boolean,nodeProps:Function,dropdownPlacement:{type:String,default:"bottom"},responsive:Boolean,items:Array,onOpenNamesChange:[Function,Array],onSelect:[Function,Array],onExpandedNamesChange:[Function,Array],expandedNames:Array,defaultExpandedNames:Array}),st=E({name:"Menu",inheritAttrs:!1,props:at,setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:t}=te(e),a=j("Menu","-menu",Qo,bo,e,o),i=V(Me,null),l=g(()=>{var p;const{collapsed:I}=e;if(I!==void 0)return I;if(i){const{collapseModeRef:r,collapsedRef:b}=i;if(r.value==="width")return(p=b.value)!==null&&p!==void 0?p:!1}return!1}),v=g(()=>{const{keyField:p,childrenField:I,disabledField:r}=e;return me(e.items||e.options,{getIgnored(b){return pe(b)},getChildren(b){return b[I]},getDisabled(b){return b[r]},getKey(b){var T;return(T=b[p])!==null&&T!==void 0?T:b.name}})}),m=g(()=>new Set(v.value.treeNodes.map(p=>p.key))),{watchProps:s}=e,h=L(null);s!=null&&s.includes("defaultValue")?ke(()=>{h.value=e.defaultValue}):h.value=e.defaultValue;const _=oe(e,"value"),N=fe(_,h),f=L([]),z=()=>{f.value=e.defaultExpandAll?v.value.getNonLeafKeys():e.defaultExpandedNames||e.defaultExpandedKeys||v.value.getPath(N.value,{includeSelf:!1}).keyPath};s!=null&&s.includes("defaultExpandedKeys")?ke(z):z();const S=Ao(e,["expandedNames","expandedKeys"]),C=fe(S,f),y=g(()=>v.value.treeNodes),A=g(()=>v.value.getPath(N.value).keyPath);Y(ne,{props:e,mergedCollapsedRef:l,mergedThemeRef:a,mergedValueRef:N,mergedExpandedKeysRef:C,activePathRef:A,mergedClsPrefixRef:o,isHorizontalRef:g(()=>e.mode==="horizontal"),invertedRef:oe(e,"inverted"),doSelect:J,toggleExpand:F});function J(p,I){const{"onUpdate:value":r,onUpdateValue:b,onSelect:T}=e;b&&M(b,p,I),r&&M(r,p,I),T&&M(T,p,I),h.value=p}function D(p){const{"onUpdate:expandedKeys":I,onUpdateExpandedKeys:r,onExpandedNamesChange:b,onOpenNamesChange:T}=e;I&&M(I,p),r&&M(r,p),b&&M(b,p),T&&M(T,p),f.value=p}function F(p){const I=Array.from(C.value),r=I.findIndex(b=>b===p);if(~r)I.splice(r,1);else{if(e.accordion&&m.value.has(p)){const b=I.findIndex(T=>m.value.has(T));b>-1&&I.splice(b,1)}I.push(p)}D(I)}const B=p=>{const I=v.value.getPath(p??N.value,{includeSelf:!1}).keyPath;if(!I.length)return;const r=Array.from(C.value),b=new Set([...r,...I]);e.accordion&&m.value.forEach(T=>{b.has(T)&&!I.includes(T)&&b.delete(T)}),D(Array.from(b))},R=g(()=>{const{inverted:p}=e,{common:{cubicBezierEaseInOut:I},self:r}=a.value,{borderRadius:b,borderColorHorizontal:T,fontSize:oo,itemHeight:to,dividerColor:ro}=r,n={"--n-divider-color":ro,"--n-bezier":I,"--n-font-size":oo,"--n-border-color-horizontal":T,"--n-border-radius":b,"--n-item-height":to};return p?(n["--n-group-text-color"]=r.groupTextColorInverted,n["--n-color"]=r.colorInverted,n["--n-item-text-color"]=r.itemTextColorInverted,n["--n-item-text-color-hover"]=r.itemTextColorHoverInverted,n["--n-item-text-color-active"]=r.itemTextColorActiveInverted,n["--n-item-text-color-child-active"]=r.itemTextColorChildActiveInverted,n["--n-item-text-color-child-active-hover"]=r.itemTextColorChildActiveInverted,n["--n-item-text-color-active-hover"]=r.itemTextColorActiveHoverInverted,n["--n-item-icon-color"]=r.itemIconColorInverted,n["--n-item-icon-color-hover"]=r.itemIconColorHoverInverted,n["--n-item-icon-color-active"]=r.itemIconColorActiveInverted,n["--n-item-icon-color-active-hover"]=r.itemIconColorActiveHoverInverted,n["--n-item-icon-color-child-active"]=r.itemIconColorChildActiveInverted,n["--n-item-icon-color-child-active-hover"]=r.itemIconColorChildActiveHoverInverted,n["--n-item-icon-color-collapsed"]=r.itemIconColorCollapsedInverted,n["--n-item-text-color-horizontal"]=r.itemTextColorHorizontalInverted,n["--n-item-text-color-hover-horizontal"]=r.itemTextColorHoverHorizontalInverted,n["--n-item-text-color-active-horizontal"]=r.itemTextColorActiveHorizontalInverted,n["--n-item-text-color-child-active-horizontal"]=r.itemTextColorChildActiveHorizontalInverted,n["--n-item-text-color-child-active-hover-horizontal"]=r.itemTextColorChildActiveHoverHorizontalInverted,n["--n-item-text-color-active-hover-horizontal"]=r.itemTextColorActiveHoverHorizontalInverted,n["--n-item-icon-color-horizontal"]=r.itemIconColorHorizontalInverted,n["--n-item-icon-color-hover-horizontal"]=r.itemIconColorHoverHorizontalInverted,n["--n-item-icon-color-active-horizontal"]=r.itemIconColorActiveHorizontalInverted,n["--n-item-icon-color-active-hover-horizontal"]=r.itemIconColorActiveHoverHorizontalInverted,n["--n-item-icon-color-child-active-horizontal"]=r.itemIconColorChildActiveHorizontalInverted,n["--n-item-icon-color-child-active-hover-horizontal"]=r.itemIconColorChildActiveHoverHorizontalInverted,n["--n-arrow-color"]=r.arrowColorInverted,n["--n-arrow-color-hover"]=r.arrowColorHoverInverted,n["--n-arrow-color-active"]=r.arrowColorActiveInverted,n["--n-arrow-color-active-hover"]=r.arrowColorActiveHoverInverted,n["--n-arrow-color-child-active"]=r.arrowColorChildActiveInverted,n["--n-arrow-color-child-active-hover"]=r.arrowColorChildActiveHoverInverted,n["--n-item-color-hover"]=r.itemColorHoverInverted,n["--n-item-color-active"]=r.itemColorActiveInverted,n["--n-item-color-active-hover"]=r.itemColorActiveHoverInverted,n["--n-item-color-active-collapsed"]=r.itemColorActiveCollapsedInverted):(n["--n-group-text-color"]=r.groupTextColor,n["--n-color"]=r.color,n["--n-item-text-color"]=r.itemTextColor,n["--n-item-text-color-hover"]=r.itemTextColorHover,n["--n-item-text-color-active"]=r.itemTextColorActive,n["--n-item-text-color-child-active"]=r.itemTextColorChildActive,n["--n-item-text-color-child-active-hover"]=r.itemTextColorChildActiveHover,n["--n-item-text-color-active-hover"]=r.itemTextColorActiveHover,n["--n-item-icon-color"]=r.itemIconColor,n["--n-item-icon-color-hover"]=r.itemIconColorHover,n["--n-item-icon-color-active"]=r.itemIconColorActive,n["--n-item-icon-color-active-hover"]=r.itemIconColorActiveHover,n["--n-item-icon-color-child-active"]=r.itemIconColorChildActive,n["--n-item-icon-color-child-active-hover"]=r.itemIconColorChildActiveHover,n["--n-item-icon-color-collapsed"]=r.itemIconColorCollapsed,n["--n-item-text-color-horizontal"]=r.itemTextColorHorizontal,n["--n-item-text-color-hover-horizontal"]=r.itemTextColorHoverHorizontal,n["--n-item-text-color-active-horizontal"]=r.itemTextColorActiveHorizontal,n["--n-item-text-color-child-active-horizontal"]=r.itemTextColorChildActiveHorizontal,n["--n-item-text-color-child-active-hover-horizontal"]=r.itemTextColorChildActiveHoverHorizontal,n["--n-item-text-color-active-hover-horizontal"]=r.itemTextColorActiveHoverHorizontal,n["--n-item-icon-color-horizontal"]=r.itemIconColorHorizontal,n["--n-item-icon-color-hover-horizontal"]=r.itemIconColorHoverHorizontal,n["--n-item-icon-color-active-horizontal"]=r.itemIconColorActiveHorizontal,n["--n-item-icon-color-active-hover-horizontal"]=r.itemIconColorActiveHoverHorizontal,n["--n-item-icon-color-child-active-horizontal"]=r.itemIconColorChildActiveHorizontal,n["--n-item-icon-color-child-active-hover-horizontal"]=r.itemIconColorChildActiveHoverHorizontal,n["--n-arrow-color"]=r.arrowColor,n["--n-arrow-color-hover"]=r.arrowColorHover,n["--n-arrow-color-active"]=r.arrowColorActive,n["--n-arrow-color-active-hover"]=r.arrowColorActiveHover,n["--n-arrow-color-child-active"]=r.arrowColorChildActive,n["--n-arrow-color-child-active-hover"]=r.arrowColorChildActiveHover,n["--n-item-color-hover"]=r.itemColorHover,n["--n-item-color-active"]=r.itemColorActive,n["--n-item-color-active-hover"]=r.itemColorActiveHover,n["--n-item-color-active-collapsed"]=r.itemColorActiveCollapsed),n}),k=t?re("menu",g(()=>e.inverted?"a":"b"),R,e):void 0,U=po(),K=L(null),se=L(null);let $=!0;const we=()=>{var p;$?$=!1:(p=K.value)===null||p===void 0||p.sync({showAllItemsBeforeCalculate:!0})};function Ye(){return document.getElementById(U)}const le=L(-1);function We(p){le.value=e.options.length-p}function Xe(p){p||(le.value=-1)}const Ze=g(()=>{const p=le.value;return{children:p===-1?[]:e.options.slice(p)}}),Je=g(()=>{const{childrenField:p,disabledField:I,keyField:r}=e;return me([Ze.value],{getIgnored(b){return pe(b)},getChildren(b){return b[p]},getDisabled(b){return b[I]},getKey(b){var T;return(T=b[r])!==null&&T!==void 0?T:b.name}})}),Qe=g(()=>me([{}]).treeNodes[0]);function eo(){var p;if(le.value===-1)return u(ge,{root:!0,level:0,key:"__ellpisisGroupPlaceholder__",internalKey:"__ellpisisGroupPlaceholder__",title:"···",tmNode:Qe.value,domId:U,isEllipsisPlaceholder:!0});const I=Je.value.treeNodes[0],r=A.value,b=!!(!((p=I.children)===null||p===void 0)&&p.some(T=>r.includes(T.key)));return u(ge,{level:0,root:!0,key:"__ellpisisGroup__",internalKey:"__ellpisisGroup__",title:"···",virtualChildActive:b,tmNode:I,domId:U,rawNodes:I.rawNode.children||[],tmNodes:I.children||[],isEllipsisPlaceholder:!0})}return{mergedClsPrefix:o,controlledExpandedKeys:S,uncontrolledExpanededKeys:f,mergedExpandedKeys:C,uncontrolledValue:h,mergedValue:N,activePath:A,tmNodes:y,mergedTheme:a,mergedCollapsed:l,cssVars:t?void 0:R,themeClass:k==null?void 0:k.themeClass,overflowRef:K,counterRef:se,updateCounter:()=>{},onResize:we,onUpdateOverflow:Xe,onUpdateCount:We,renderCounter:eo,getCounter:Ye,onRender:k==null?void 0:k.onRender,showOption:B,deriveResponsiveState:we}},render(){const{mergedClsPrefix:e,mode:o,themeClass:t,onRender:a}=this;a==null||a();const i=()=>this.tmNodes.map(s=>Ie(s,this.$props)),v=o==="horizontal"&&this.responsive,m=()=>u("div",go(this.$attrs,{role:o==="horizontal"?"menubar":"menu",class:[`${e}-menu`,t,`${e}-menu--${o}`,v&&`${e}-menu--responsive`,this.mergedCollapsed&&`${e}-menu--collapsed`],style:this.cssVars}),v?u(To,{ref:"overflowRef",onUpdateOverflow:this.onUpdateOverflow,getCounter:this.getCounter,onUpdateCount:this.onUpdateCount,updateCounter:this.updateCounter,style:{width:"100%",display:"flex",overflow:"hidden"}},{default:i,counter:this.renderCounter}):i());return v?u(fo,{onResize:this.onResize},{default:m}):m()}}),ct={class:"h-14 flex items-center justify-center px-4 border-b border-border"},dt={key:1,class:"flex-1"},Rt=E({__name:"AppLayout",setup(e){const{t:o}=xo(),t=wo(),a=ko(),i=Co(),l=yo(),v=g(()=>[{label:o("nav.dashboard"),key:"dashboard"},{label:o("nav.users"),key:"users"},{label:o("nav.groups"),key:"groups"},{label:o("nav.channels"),key:"channels"},{label:o("nav.accounts"),key:"accounts"},{label:o("nav.models"),key:"models"},{label:o("nav.tokens"),key:"tokens"},{label:o("nav.logs"),key:"logs",children:[{label:o("nav.logs_requests"),key:"logs-requests"},{label:o("nav.logs_errors"),key:"logs-errors"},{label:o("nav.logs_audit"),key:"logs-audit"}]},{label:o("nav.stats"),key:"stats"},{label:o("nav.notices"),key:"notices"},{label:o("nav.payments"),key:"payments",children:[{label:o("nav.payments_recharges"),key:"payments-recharges"},{label:o("nav.payments_redeem"),key:"payments-redeem"}]},{label:o("nav.settings"),key:"settings"}]);function m(z){a.hasRoute(z)&&a.push({name:z})}const s={"channel-detail":"channels","channel-new":"channels","channel-mappings":"channels","user-detail":"users","account-detail":"accounts","notice-detail":"notices","notice-new":"notices","payments-redeem-batch":"payments-redeem"},h=g(()=>{const z=String(t.name??"dashboard");return s[z]??z}),_=[{label:o("nav.profile"),key:"profile"},{label:o("nav.logout"),key:"logout"}];async function N(z){z==="logout"?(await l.logout(),a.push({name:"login"})):z==="profile"&&a.push({name:"profile"})}const f=g(()=>{var S;const z=(S=t.meta)==null?void 0:S.breadcrumb;return z?z.map(C=>({label:C.startsWith(":")?t.params[C.slice(1)]:o(`nav.${C}`),key:C})):[]});return(z,S)=>{const C=Ro("RouterView");return G(),W(w(Ae),{"has-sider":"",style:{height:"100vh"}},{default:H(()=>[O(w(Jo),{collapsed:w(i).sidebarCollapsed,"collapse-mode":"width","collapsed-width":64,width:220,bordered:"","show-trigger":"bar","onUpdate:collapsed":w(i).toggleSidebar},{default:H(()=>[de("div",ct,[w(i).sidebarCollapsed?(G(),W(w(Ne),{key:1,strong:""},{default:H(()=>[...S[2]||(S[2]=[ee("P",-1)])]),_:1})):(G(),W(w(Ne),{key:0,strong:"",class:"text-lg"},{default:H(()=>[...S[1]||(S[1]=[ee("pro-api",-1)])]),_:1}))]),O(w(st),{collapsed:w(i).sidebarCollapsed,"collapsed-width":64,"collapsed-icon-size":20,options:v.value,value:h.value,"default-expanded-keys":["logs","payments"],"onUpdate:value":m},null,8,["collapsed","options","value"])]),_:1},8,["collapsed","onUpdate:collapsed"]),O(w(Ae),null,{default:H(()=>[O(w(qo),{bordered:"",class:"px-4 h-14 flex items-center gap-3"},{default:H(()=>[O(w(ie),{text:"",onClick:w(i).toggleSidebar},{default:H(()=>[O(w(Te),{size:"20"},{default:H(()=>[...S[3]||(S[3]=[de("span",{class:"i-lucide-menu"},null,-1)])]),_:1})]),_:1},8,["onClick"]),f.value.length?(G(),W(w($o),{key:0,class:"flex-1"},{default:H(()=>[(G(!0),Pe($e,null,zo(f.value,y=>(G(),W(w(Mo),{key:y.key},{default:H(()=>[ee(ue(y.label),1)]),_:2},1024))),128))]),_:1})):(G(),Pe("div",dt)),O(w(_o),{align:"center",size:12},{default:H(()=>[O(w(ie),{text:"",onClick:w(i).toggleTheme},{default:H(()=>[O(w(Te),{size:"20"},{default:H(()=>[de("span",{class:So(w(i).resolvedTheme==="dark"?"i-lucide-sun":"i-lucide-moon")},null,2)]),_:1})]),_:1},8,["onClick"]),O(w(ie),{text:"",onClick:S[0]||(S[0]=y=>w(i).setLocale(w(i).locale==="zh"?"en":"zh")),size:"small"},{default:H(()=>[ee(ue(w(i).locale==="zh"?"EN":"中"),1)]),_:1}),O(w(Le),{options:_,onSelect:N},{default:H(()=>[O(w(ie),{text:""},{default:H(()=>[O(w(Bo),{type:"primary",size:"small"},{default:H(()=>{var y;return[ee(ue(((y=w(l).user)==null?void 0:y.username)??"--"),1)]}),_:1})]),_:1})]),_:1})]),_:1})]),_:1}),O(w(Do),{class:"p-4 overflow-auto",style:{height:"calc(100vh - 56px)"}},{default:H(()=>[O(C,null,{default:H(({Component:y})=>[(G(),W(Io(y),{key:w(t).fullPath}))]),_:1})]),_:1})]),_:1})]),_:1})}}});export{Rt as default};
