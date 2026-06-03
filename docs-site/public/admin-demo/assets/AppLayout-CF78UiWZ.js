import{al as L,aU as i,y as u,x as b,D as P,A as v,cu as re,cI as K,cJ as ne,v as lo,a3 as x,ab as J,bU as W,cn as te,bZ as F,bI as io,bJ as ao,b6 as so,aZ as D,c6 as co,af as uo,cc as vo,ap as mo,a2 as ke,S as Oe,cC as Ee,d as Le,H as j,G as ee,aD as ho,b_ as Z,bl as xe,cz as fe,b as $e,bk as de,g as fo,V as po,cO as Pe,aa as go,bp as bo,bo as xo,cw as Co,ct as yo,cK as zo,a7 as X,cP as H,cs as S,R as O,bK as q,ag as E,a6 as ue,ae as oe,B as ae,a9 as Ne,b$ as Io,by as So,cm as ve,m as wo,c5 as Ro,c4 as ko,cE as Po,cD as No}from"./index-ClKeNBH9.js";import{N as Te}from"./text-CiDnAQYY.js";import{N as To}from"./Tooltip-jxqJeGHW.js";import{C as _o,N as Fe}from"./Dropdown-sI2kx75c.js";import{f as me,u as pe}from"./get-48VdzrSm.js";import{V as Ao,a as he}from"./create-DcIarVxf.js";import{u as Bo}from"./cssr-DxXR4Bge.js";import{N as _e}from"./Icon-jpXWQhB8.js";import{N as Ho}from"./Space-C_SQi-yO.js";import{N as Oo}from"./Tag-BnfjH-Aw.js";import{_ as Eo}from"./_plugin-vue_export-helper-DlAUqK2U.js";import"./Popover-Dm198j0w.js";import"./happens-in-CM8LO42l.js";import"./use-keyboard-C3dMSzsu.js";import"./create-ref-setter-C4J8sofl.js";import"./get-slot-Bk_rJcZu.js";const Lo=L({name:"ChevronDownFilled",render(){return i("svg",{viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg"},i("path",{d:"M3.20041 5.73966C3.48226 5.43613 3.95681 5.41856 4.26034 5.70041L8 9.22652L11.7397 5.70041C12.0432 5.41856 12.5177 5.43613 12.7996 5.73966C13.0815 6.0432 13.0639 6.51775 12.7603 6.7996L8.51034 10.7996C8.22258 11.0668 7.77743 11.0668 7.48967 10.7996L3.23966 6.7996C2.93613 6.51775 2.91856 6.0432 3.20041 5.73966Z",fill:"currentColor"}))}}),$o=u("breadcrumb",`
 white-space: nowrap;
 cursor: default;
 line-height: var(--n-item-line-height);
`,[b("ul",`
 list-style: none;
 padding: 0;
 margin: 0;
 `),b("a",`
 color: inherit;
 text-decoration: inherit;
 `),u("breadcrumb-item",`
 font-size: var(--n-font-size);
 transition: color .3s var(--n-bezier);
 display: inline-flex;
 align-items: center;
 `,[u("icon",`
 font-size: 18px;
 vertical-align: -.2em;
 transition: color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 `),b("&:not(:last-child)",[P("clickable",[v("link",`
 cursor: pointer;
 `,[b("&:hover",`
 background-color: var(--n-item-color-hover);
 `),b("&:active",`
 background-color: var(--n-item-color-pressed); 
 `)])])]),v("link",`
 padding: 4px;
 border-radius: var(--n-item-border-radius);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 position: relative;
 `,[b("&:hover",`
 color: var(--n-item-text-color-hover);
 `,[u("icon",`
 color: var(--n-item-text-color-hover);
 `)]),b("&:active",`
 color: var(--n-item-text-color-pressed);
 `,[u("icon",`
 color: var(--n-item-text-color-pressed);
 `)])]),v("separator",`
 margin: 0 8px;
 color: var(--n-separator-color);
 transition: color .3s var(--n-bezier);
 user-select: none;
 -webkit-user-select: none;
 `),b("&:last-child",[v("link",`
 font-weight: var(--n-font-weight-active);
 cursor: unset;
 color: var(--n-item-text-color-active);
 `,[u("icon",`
 color: var(--n-item-text-color-active);
 `)]),v("separator",`
 display: none;
 `)])])]),Me=J("n-breadcrumb"),Fo=Object.assign(Object.assign({},K.props),{separator:{type:String,default:"/"}}),Mo=L({name:"Breadcrumb",props:Fo,setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:t}=re(e),s=K("Breadcrumb","-breadcrumb",$o,lo,e,o);W(Me,{separatorRef:te(e,"separator"),mergedClsPrefixRef:o});const a=x(()=>{const{common:{cubicBezierEaseInOut:m},self:{separatorColor:h,itemTextColor:c,itemTextColorHover:f,itemTextColorPressed:_,itemTextColorActive:w,fontSize:d,fontWeightActive:k,itemBorderRadius:y,itemColorHover:R,itemColorPressed:N,itemLineHeight:A}}=s.value;return{"--n-font-size":d,"--n-bezier":m,"--n-item-text-color":c,"--n-item-text-color-hover":f,"--n-item-text-color-pressed":_,"--n-item-text-color-active":w,"--n-separator-color":h,"--n-item-color-hover":R,"--n-item-color-pressed":N,"--n-item-border-radius":y,"--n-font-weight-active":k,"--n-item-line-height":A}}),l=t?ne("breadcrumb",void 0,a,e):void 0;return{mergedClsPrefix:o,cssVars:t?void 0:a,themeClass:l==null?void 0:l.themeClass,onRender:l==null?void 0:l.onRender}},render(){var e;return(e=this.onRender)===null||e===void 0||e.call(this),i("nav",{class:[`${this.mergedClsPrefix}-breadcrumb`,this.themeClass],style:this.cssVars,"aria-label":"Breadcrumb"},i("ul",null,this.$slots))}});function jo(e=so?window:null){const o=()=>{const{hash:a,host:l,hostname:m,href:h,origin:c,pathname:f,port:_,protocol:w,search:d}=(e==null?void 0:e.location)||{};return{hash:a,host:l,hostname:m,href:h,origin:c,pathname:f,port:_,protocol:w,search:d}},t=F(o()),s=()=>{t.value=o()};return io(()=>{e&&(e.addEventListener("popstate",s),e.addEventListener("hashchange",s))}),ao(()=>{e&&(e.removeEventListener("popstate",s),e.removeEventListener("hashchange",s))}),t}const Ko={separator:String,href:String,clickable:{type:Boolean,default:!0},showSeparator:{type:Boolean,default:!0},onClick:Function},Vo=L({name:"BreadcrumbItem",props:Ko,slots:Object,setup(e,{slots:o}){const t=D(Me,null);if(!t)return()=>null;const{separatorRef:s,mergedClsPrefixRef:a}=t,l=jo(),m=x(()=>e.href?"a":"span"),h=x(()=>l.value.href===e.href?"location":null);return()=>{const{value:c}=a;return i("li",{class:[`${c}-breadcrumb-item`,e.clickable&&`${c}-breadcrumb-item--clickable`]},i(m.value,{class:`${c}-breadcrumb-item__link`,"aria-current":h.value,href:e.href,onClick:e.onClick},o),e.showSeparator&&i("span",{class:`${c}-breadcrumb-item__separator`,"aria-hidden":"true"},co(o.separator,()=>{var f;return[(f=e.separator)!==null&&f!==void 0?f:s.value]})))}}});function Do(e){const{baseColor:o,textColor2:t,bodyColor:s,cardColor:a,dividerColor:l,actionColor:m,scrollbarColor:h,scrollbarColorHover:c,invertedColor:f}=e;return{textColor:t,textColorInverted:"#FFF",color:s,colorEmbedded:m,headerColor:a,headerColorInverted:f,footerColor:m,footerColorInverted:f,headerBorderColor:l,headerBorderColorInverted:f,footerBorderColor:l,footerBorderColorInverted:f,siderBorderColor:l,siderBorderColorInverted:f,siderColor:a,siderColorInverted:f,siderToggleButtonBorder:`1px solid ${l}`,siderToggleButtonColor:o,siderToggleButtonIconColor:t,siderToggleButtonIconColorInverted:t,siderToggleBarColor:ke(s,h),siderToggleBarColorHover:ke(s,c),__invertScrollbar:"true"}}const Ce=uo({name:"Layout",common:mo,peers:{Scrollbar:vo},self:Do}),je=J("n-layout-sider"),ye={type:String,default:"static"},Uo=u("layout",`
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
`,[u("layout-scroll-container",`
 overflow-x: hidden;
 box-sizing: border-box;
 height: 100%;
 `),P("absolute-positioned",`
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),Go={embedded:Boolean,position:ye,nativeScrollbar:{type:Boolean,default:!0},scrollbarProps:Object,onScroll:Function,contentClass:String,contentStyle:{type:[String,Object],default:""},hasSider:Boolean,siderPlacement:{type:String,default:"left"}},Ke=J("n-layout");function Ve(e){return L({name:e?"LayoutContent":"Layout",props:Object.assign(Object.assign({},K.props),Go),setup(o){const t=F(null),s=F(null),{mergedClsPrefixRef:a,inlineThemeDisabled:l}=re(o),m=K("Layout","-layout",Uo,Ce,o,a);function h(R,N){if(o.nativeScrollbar){const{value:A}=t;A&&(N===void 0?A.scrollTo(R):A.scrollTo(R,N))}else{const{value:A}=s;A&&A.scrollTo(R,N)}}W(Ke,o);let c=0,f=0;const _=R=>{var N;const A=R.target;c=A.scrollLeft,f=A.scrollTop,(N=o.onScroll)===null||N===void 0||N.call(o,R)};Ee(()=>{if(o.nativeScrollbar){const R=t.value;R&&(R.scrollTop=f,R.scrollLeft=c)}});const w={display:"flex",flexWrap:"nowrap",width:"100%",flexDirection:"row"},d={scrollTo:h},k=x(()=>{const{common:{cubicBezierEaseInOut:R},self:N}=m.value;return{"--n-bezier":R,"--n-color":o.embedded?N.colorEmbedded:N.color,"--n-text-color":N.textColor}}),y=l?ne("layout",x(()=>o.embedded?"e":""),k,o):void 0;return Object.assign({mergedClsPrefix:a,scrollableElRef:t,scrollbarInstRef:s,hasSiderStyle:w,mergedTheme:m,handleNativeElScroll:_,cssVars:l?void 0:k,themeClass:y==null?void 0:y.themeClass,onRender:y==null?void 0:y.onRender},d)},render(){var o;const{mergedClsPrefix:t,hasSider:s}=this;(o=this.onRender)===null||o===void 0||o.call(this);const a=s?this.hasSiderStyle:void 0,l=[this.themeClass,e&&`${t}-layout-content`,`${t}-layout`,`${t}-layout--${this.position}-positioned`];return i("div",{class:l,style:this.cssVars},this.nativeScrollbar?i("div",{ref:"scrollableElRef",class:[`${t}-layout-scroll-container`,this.contentClass],style:[this.contentStyle,a],onScroll:this.handleNativeElScroll},this.$slots):i(Oe,Object.assign({},this.scrollbarProps,{onScroll:this.onScroll,ref:"scrollbarInstRef",theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,contentClass:this.contentClass,contentStyle:[this.contentStyle,a]}),this.$slots))}})}const Ae=Ve(!1),qo=Ve(!0),Yo=u("layout-header",`
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
 `)]),Wo={position:ye,inverted:Boolean,bordered:{type:Boolean,default:!1}},Xo=L({name:"LayoutHeader",props:Object.assign(Object.assign({},K.props),Wo),setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:t}=re(e),s=K("Layout","-layout-header",Yo,Ce,e,o),a=x(()=>{const{common:{cubicBezierEaseInOut:m},self:h}=s.value,c={"--n-bezier":m};return e.inverted?(c["--n-color"]=h.headerColorInverted,c["--n-text-color"]=h.textColorInverted,c["--n-border-color"]=h.headerBorderColorInverted):(c["--n-color"]=h.headerColor,c["--n-text-color"]=h.textColor,c["--n-border-color"]=h.headerBorderColor),c}),l=t?ne("layout-header",x(()=>e.inverted?"a":"b"),a,e):void 0;return{mergedClsPrefix:o,cssVars:t?void 0:a,themeClass:l==null?void 0:l.themeClass,onRender:l==null?void 0:l.onRender}},render(){var e;const{mergedClsPrefix:o}=this;return(e=this.onRender)===null||e===void 0||e.call(this),i("div",{class:[`${o}-layout-header`,this.themeClass,this.position&&`${o}-layout-header--${this.position}-positioned`,this.bordered&&`${o}-layout-header--bordered`],style:this.cssVars},this.$slots)}}),Zo=u("layout-sider",`
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
`,[P("bordered",[v("border",`
 content: "";
 position: absolute;
 top: 0;
 bottom: 0;
 width: 1px;
 background-color: var(--n-border-color);
 transition: background-color .3s var(--n-bezier);
 `)]),v("left-placement",[P("bordered",[v("border",`
 right: 0;
 `)])]),P("right-placement",`
 justify-content: flex-start;
 `,[P("bordered",[v("border",`
 left: 0;
 `)]),P("collapsed",[u("layout-toggle-button",[u("base-icon",`
 transform: rotate(180deg);
 `)]),u("layout-toggle-bar",[b("&:hover",[v("top",{transform:"rotate(-12deg) scale(1.15) translateY(-2px)"}),v("bottom",{transform:"rotate(12deg) scale(1.15) translateY(2px)"})])])]),u("layout-toggle-button",`
 left: 0;
 transform: translateX(-50%) translateY(-50%);
 `,[u("base-icon",`
 transform: rotate(0);
 `)]),u("layout-toggle-bar",`
 left: -28px;
 transform: rotate(180deg);
 `,[b("&:hover",[v("top",{transform:"rotate(12deg) scale(1.15) translateY(-2px)"}),v("bottom",{transform:"rotate(-12deg) scale(1.15) translateY(2px)"})])])]),P("collapsed",[u("layout-toggle-bar",[b("&:hover",[v("top",{transform:"rotate(-12deg) scale(1.15) translateY(-2px)"}),v("bottom",{transform:"rotate(12deg) scale(1.15) translateY(2px)"})])]),u("layout-toggle-button",[u("base-icon",`
 transform: rotate(0);
 `)])]),u("layout-toggle-button",`
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
 `,[u("base-icon",`
 transition: transform .3s var(--n-bezier);
 transform: rotate(180deg);
 `)]),u("layout-toggle-bar",`
 cursor: pointer;
 height: 72px;
 width: 32px;
 position: absolute;
 top: calc(50% - 36px);
 right: -28px;
 `,[v("top, bottom",`
 position: absolute;
 width: 4px;
 border-radius: 2px;
 height: 38px;
 left: 14px;
 transition: 
 background-color .3s var(--n-bezier),
 transform .3s var(--n-bezier);
 `),v("bottom",`
 position: absolute;
 top: 34px;
 `),b("&:hover",[v("top",{transform:"rotate(12deg) scale(1.15) translateY(-2px)"}),v("bottom",{transform:"rotate(-12deg) scale(1.15) translateY(2px)"})]),v("top, bottom",{backgroundColor:"var(--n-toggle-bar-color)"}),b("&:hover",[v("top, bottom",{backgroundColor:"var(--n-toggle-bar-color-hover)"})])]),v("border",`
 position: absolute;
 top: 0;
 right: 0;
 bottom: 0;
 width: 1px;
 transition: background-color .3s var(--n-bezier);
 `),u("layout-sider-scroll-container",`
 flex-grow: 1;
 flex-shrink: 0;
 box-sizing: border-box;
 height: 100%;
 opacity: 0;
 transition: opacity .3s var(--n-bezier);
 max-width: 100%;
 `),P("show-content",[u("layout-sider-scroll-container",{opacity:1})]),P("absolute-positioned",`
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 `)]),Jo=L({props:{clsPrefix:{type:String,required:!0},onClick:Function},render(){const{clsPrefix:e}=this;return i("div",{onClick:this.onClick,class:`${e}-layout-toggle-bar`},i("div",{class:`${e}-layout-toggle-bar__top`}),i("div",{class:`${e}-layout-toggle-bar__bottom`}))}}),Qo=L({name:"LayoutToggleButton",props:{clsPrefix:{type:String,required:!0},onClick:Function},render(){const{clsPrefix:e}=this;return i("div",{class:`${e}-layout-toggle-button`,onClick:this.onClick},i(Le,{clsPrefix:e},{default:()=>i(_o,null)}))}}),et={position:ye,bordered:Boolean,collapsedWidth:{type:Number,default:48},width:{type:[Number,String],default:272},contentClass:String,contentStyle:{type:[String,Object],default:""},collapseMode:{type:String,default:"transform"},collapsed:{type:Boolean,default:void 0},defaultCollapsed:Boolean,showCollapsedContent:{type:Boolean,default:!0},showTrigger:{type:[Boolean,String],default:!1},nativeScrollbar:{type:Boolean,default:!0},inverted:Boolean,scrollbarProps:Object,triggerClass:String,triggerStyle:[String,Object],collapsedTriggerClass:String,collapsedTriggerStyle:[String,Object],"onUpdate:collapsed":[Function,Array],onUpdateCollapsed:[Function,Array],onAfterEnter:Function,onAfterLeave:Function,onExpand:[Function,Array],onCollapse:[Function,Array],onScroll:Function},ot=L({name:"LayoutSider",props:Object.assign(Object.assign({},K.props),et),setup(e){const o=D(Ke),t=F(null),s=F(null),a=F(e.defaultCollapsed),l=pe(te(e,"collapsed"),a),m=x(()=>me(l.value?e.collapsedWidth:e.width)),h=x(()=>e.collapseMode!=="transform"?{}:{minWidth:me(e.width)}),c=x(()=>o?o.siderPlacement:"left");function f(B,z){if(e.nativeScrollbar){const{value:I}=t;I&&(z===void 0?I.scrollTo(B):I.scrollTo(B,z))}else{const{value:I}=s;I&&I.scrollTo(B,z)}}function _(){const{"onUpdate:collapsed":B,onUpdateCollapsed:z,onExpand:I,onCollapse:G}=e,{value:V}=l;z&&j(z,!V),B&&j(B,!V),a.value=!V,V?I&&j(I):G&&j(G)}let w=0,d=0;const k=B=>{var z;const I=B.target;w=I.scrollLeft,d=I.scrollTop,(z=e.onScroll)===null||z===void 0||z.call(e,B)};Ee(()=>{if(e.nativeScrollbar){const B=t.value;B&&(B.scrollTop=d,B.scrollLeft=w)}}),W(je,{collapsedRef:l,collapseModeRef:te(e,"collapseMode")});const{mergedClsPrefixRef:y,inlineThemeDisabled:R}=re(e),N=K("Layout","-layout-sider",Zo,Ce,e,y);function A(B){var z,I;B.propertyName==="max-width"&&(l.value?(z=e.onAfterLeave)===null||z===void 0||z.call(e):(I=e.onAfterEnter)===null||I===void 0||I.call(e))}const Q={scrollTo:f},U=x(()=>{const{common:{cubicBezierEaseInOut:B},self:z}=N.value,{siderToggleButtonColor:I,siderToggleButtonBorder:G,siderToggleBarColor:V,siderToggleBarColorHover:ce}=z,$={"--n-bezier":B,"--n-toggle-button-color":I,"--n-toggle-button-border":G,"--n-toggle-bar-color":V,"--n-toggle-bar-color-hover":ce};return e.inverted?($["--n-color"]=z.siderColorInverted,$["--n-text-color"]=z.textColorInverted,$["--n-border-color"]=z.siderBorderColorInverted,$["--n-toggle-button-icon-color"]=z.siderToggleButtonIconColorInverted,$.__invertScrollbar=z.__invertScrollbar):($["--n-color"]=z.siderColor,$["--n-text-color"]=z.textColor,$["--n-border-color"]=z.siderBorderColor,$["--n-toggle-button-icon-color"]=z.siderToggleButtonIconColor),$}),M=R?ne("layout-sider",x(()=>e.inverted?"a":"b"),U,e):void 0;return Object.assign({scrollableElRef:t,scrollbarInstRef:s,mergedClsPrefix:y,mergedTheme:N,styleMaxWidth:m,mergedCollapsed:l,scrollContainerStyle:h,siderPlacement:c,handleNativeElScroll:k,handleTransitionend:A,handleTriggerClick:_,inlineThemeDisabled:R,cssVars:U,themeClass:M==null?void 0:M.themeClass,onRender:M==null?void 0:M.onRender},Q)},render(){var e;const{mergedClsPrefix:o,mergedCollapsed:t,showTrigger:s}=this;return(e=this.onRender)===null||e===void 0||e.call(this),i("aside",{class:[`${o}-layout-sider`,this.themeClass,`${o}-layout-sider--${this.position}-positioned`,`${o}-layout-sider--${this.siderPlacement}-placement`,this.bordered&&`${o}-layout-sider--bordered`,t&&`${o}-layout-sider--collapsed`,(!t||this.showCollapsedContent)&&`${o}-layout-sider--show-content`],onTransitionend:this.handleTransitionend,style:[this.inlineThemeDisabled?void 0:this.cssVars,{maxWidth:this.styleMaxWidth,width:me(this.width)}]},this.nativeScrollbar?i("div",{class:[`${o}-layout-sider-scroll-container`,this.contentClass],onScroll:this.handleNativeElScroll,style:[this.scrollContainerStyle,{overflow:"auto"},this.contentStyle],ref:"scrollableElRef"},this.$slots):i(Oe,Object.assign({},this.scrollbarProps,{onScroll:this.onScroll,ref:"scrollbarInstRef",style:this.scrollContainerStyle,contentStyle:this.contentStyle,contentClass:this.contentClass,theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,builtinThemeOverrides:this.inverted&&this.cssVars.__invertScrollbar==="true"?{colorHover:"rgba(255, 255, 255, .4)",color:"rgba(255, 255, 255, .3)"}:void 0}),this.$slots),s?s==="bar"?i(Jo,{clsPrefix:o,class:t?this.collapsedTriggerClass:this.triggerClass,style:t?this.collapsedTriggerStyle:this.triggerStyle,onClick:this.handleTriggerClick}):i(Qo,{clsPrefix:o,class:t?this.collapsedTriggerClass:this.triggerClass,style:t?this.collapsedTriggerStyle:this.triggerStyle,onClick:this.handleTriggerClick}):null,this.bordered?i("div",{class:`${o}-layout-sider__border`}):null)}}),le=J("n-menu"),De=J("n-submenu"),ze=J("n-menu-item-group"),Be=[b("&::before","background-color: var(--n-item-color-hover);"),v("arrow",`
 color: var(--n-arrow-color-hover);
 `),v("icon",`
 color: var(--n-item-icon-color-hover);
 `),u("menu-item-content-header",`
 color: var(--n-item-text-color-hover);
 `,[b("a",`
 color: var(--n-item-text-color-hover);
 `),v("extra",`
 color: var(--n-item-text-color-hover);
 `)])],He=[v("icon",`
 color: var(--n-item-icon-color-hover-horizontal);
 `),u("menu-item-content-header",`
 color: var(--n-item-text-color-hover-horizontal);
 `,[b("a",`
 color: var(--n-item-text-color-hover-horizontal);
 `),v("extra",`
 color: var(--n-item-text-color-hover-horizontal);
 `)])],tt=b([u("menu",`
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
 `,[u("submenu","margin: 0;"),u("menu-item","margin: 0;"),u("menu-item-content",`
 padding: 0 20px;
 border-bottom: 2px solid #0000;
 `,[b("&::before","display: none;"),P("selected","border-bottom: 2px solid var(--n-border-color-horizontal)")]),u("menu-item-content",[P("selected",[v("icon","color: var(--n-item-icon-color-active-horizontal);"),u("menu-item-content-header",`
 color: var(--n-item-text-color-active-horizontal);
 `,[b("a","color: var(--n-item-text-color-active-horizontal);"),v("extra","color: var(--n-item-text-color-active-horizontal);")])]),P("child-active",`
 border-bottom: 2px solid var(--n-border-color-horizontal);
 `,[u("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-horizontal);
 `,[b("a",`
 color: var(--n-item-text-color-child-active-horizontal);
 `),v("extra",`
 color: var(--n-item-text-color-child-active-horizontal);
 `)]),v("icon",`
 color: var(--n-item-icon-color-child-active-horizontal);
 `)]),ee("disabled",[ee("selected, child-active",[b("&:focus-within",He)]),P("selected",[Y(null,[v("icon","color: var(--n-item-icon-color-active-hover-horizontal);"),u("menu-item-content-header",`
 color: var(--n-item-text-color-active-hover-horizontal);
 `,[b("a","color: var(--n-item-text-color-active-hover-horizontal);"),v("extra","color: var(--n-item-text-color-active-hover-horizontal);")])])]),P("child-active",[Y(null,[v("icon","color: var(--n-item-icon-color-child-active-hover-horizontal);"),u("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-hover-horizontal);
 `,[b("a","color: var(--n-item-text-color-child-active-hover-horizontal);"),v("extra","color: var(--n-item-text-color-child-active-hover-horizontal);")])])]),Y("border-bottom: 2px solid var(--n-border-color-horizontal);",He)]),u("menu-item-content-header",[b("a","color: var(--n-item-text-color-horizontal);")])])]),ee("responsive",[u("menu-item-content-header",`
 overflow: hidden;
 text-overflow: ellipsis;
 `)]),P("collapsed",[u("menu-item-content",[P("selected",[b("&::before",`
 background-color: var(--n-item-color-active-collapsed) !important;
 `)]),u("menu-item-content-header","opacity: 0;"),v("arrow","opacity: 0;"),v("icon","color: var(--n-item-icon-color-collapsed);")])]),u("menu-item",`
 height: var(--n-item-height);
 margin-top: 6px;
 position: relative;
 `),u("menu-item-content",`
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
 `,[b("> *","z-index: 1;"),b("&::before",`
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
 `),P("collapsed",[v("arrow","transform: rotate(0);")]),P("selected",[b("&::before","background-color: var(--n-item-color-active);"),v("arrow","color: var(--n-arrow-color-active);"),v("icon","color: var(--n-item-icon-color-active);"),u("menu-item-content-header",`
 color: var(--n-item-text-color-active);
 `,[b("a","color: var(--n-item-text-color-active);"),v("extra","color: var(--n-item-text-color-active);")])]),P("child-active",[u("menu-item-content-header",`
 color: var(--n-item-text-color-child-active);
 `,[b("a",`
 color: var(--n-item-text-color-child-active);
 `),v("extra",`
 color: var(--n-item-text-color-child-active);
 `)]),v("arrow",`
 color: var(--n-arrow-color-child-active);
 `),v("icon",`
 color: var(--n-item-icon-color-child-active);
 `)]),ee("disabled",[ee("selected, child-active",[b("&:focus-within",Be)]),P("selected",[Y(null,[v("arrow","color: var(--n-arrow-color-active-hover);"),v("icon","color: var(--n-item-icon-color-active-hover);"),u("menu-item-content-header",`
 color: var(--n-item-text-color-active-hover);
 `,[b("a","color: var(--n-item-text-color-active-hover);"),v("extra","color: var(--n-item-text-color-active-hover);")])])]),P("child-active",[Y(null,[v("arrow","color: var(--n-arrow-color-child-active-hover);"),v("icon","color: var(--n-item-icon-color-child-active-hover);"),u("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-hover);
 `,[b("a","color: var(--n-item-text-color-child-active-hover);"),v("extra","color: var(--n-item-text-color-child-active-hover);")])])]),P("selected",[Y(null,[b("&::before","background-color: var(--n-item-color-active-hover);")])]),Y(null,Be)]),v("icon",`
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
 `),v("arrow",`
 grid-area: arrow;
 font-size: 16px;
 color: var(--n-arrow-color);
 transform: rotate(180deg);
 opacity: 1;
 transition:
 color .3s var(--n-bezier),
 transform 0.2s var(--n-bezier),
 opacity 0.2s var(--n-bezier);
 `),u("menu-item-content-header",`
 grid-area: content;
 transition:
 color .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 opacity: 1;
 white-space: nowrap;
 color: var(--n-item-text-color);
 `,[b("a",`
 outline: none;
 text-decoration: none;
 transition: color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 `,[b("&::before",`
 content: "";
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),v("extra",`
 font-size: .93em;
 color: var(--n-group-text-color);
 transition: color .3s var(--n-bezier);
 `)])]),u("submenu",`
 cursor: pointer;
 position: relative;
 margin-top: 6px;
 `,[u("menu-item-content",`
 height: var(--n-item-height);
 `),u("submenu-children",`
 overflow: hidden;
 padding: 0;
 `,[ho({duration:".2s"})])]),u("menu-item-group",[u("menu-item-group-title",`
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
 `)])]),u("menu-tooltip",[b("a",`
 color: inherit;
 text-decoration: none;
 `)]),u("menu-divider",`
 transition: background-color .3s var(--n-bezier);
 background-color: var(--n-divider-color);
 height: 1px;
 margin: 6px 18px;
 `)]);function Y(e,o){return[P("hover",e,o),b("&:hover",e,o)]}const Ue=L({name:"MenuOptionContent",props:{collapsed:Boolean,disabled:Boolean,title:[String,Function],icon:Function,extra:[String,Function],showArrow:Boolean,childActive:Boolean,hover:Boolean,paddingLeft:Number,selected:Boolean,maxIconSize:{type:Number,required:!0},activeIconSize:{type:Number,required:!0},iconMarginRight:{type:Number,required:!0},clsPrefix:{type:String,required:!0},onClick:Function,tmNode:{type:Object,required:!0},isEllipsisPlaceholder:Boolean},setup(e){const{props:o}=D(le);return{menuProps:o,style:x(()=>{const{paddingLeft:t}=e;return{paddingLeft:t&&`${t}px`}}),iconStyle:x(()=>{const{maxIconSize:t,activeIconSize:s,iconMarginRight:a}=e;return{width:`${t}px`,height:`${t}px`,fontSize:`${s}px`,marginRight:`${a}px`}})}},render(){const{clsPrefix:e,tmNode:o,menuProps:{renderIcon:t,renderLabel:s,renderExtra:a,expandIcon:l}}=this,m=t?t(o.rawNode):Z(this.icon);return i("div",{onClick:h=>{var c;(c=this.onClick)===null||c===void 0||c.call(this,h)},role:"none",class:[`${e}-menu-item-content`,{[`${e}-menu-item-content--selected`]:this.selected,[`${e}-menu-item-content--collapsed`]:this.collapsed,[`${e}-menu-item-content--child-active`]:this.childActive,[`${e}-menu-item-content--disabled`]:this.disabled,[`${e}-menu-item-content--hover`]:this.hover}],style:this.style},m&&i("div",{class:`${e}-menu-item-content__icon`,style:this.iconStyle,role:"none"},[m]),i("div",{class:`${e}-menu-item-content-header`,role:"none"},this.isEllipsisPlaceholder?this.title:s?s(o.rawNode):Z(this.title),this.extra||a?i("span",{class:`${e}-menu-item-content-header__extra`}," ",a?a(o.rawNode):Z(this.extra)):null),this.showArrow?i(Le,{ariaHidden:!0,class:`${e}-menu-item-content__arrow`,clsPrefix:e},{default:()=>l?l(o.rawNode):i(Lo,null)}):null)}}),se=8;function Ie(e){const o=D(le),{props:t,mergedCollapsedRef:s}=o,a=D(De,null),l=D(ze,null),m=x(()=>t.mode==="horizontal"),h=x(()=>m.value?t.dropdownPlacement:"tmNodes"in e?"right-start":"right"),c=x(()=>{var d;return Math.max((d=t.collapsedIconSize)!==null&&d!==void 0?d:t.iconSize,t.iconSize)}),f=x(()=>{var d;return!m.value&&e.root&&s.value&&(d=t.collapsedIconSize)!==null&&d!==void 0?d:t.iconSize}),_=x(()=>{if(m.value)return;const{collapsedWidth:d,indent:k,rootIndent:y}=t,{root:R,isGroup:N}=e,A=y===void 0?k:y;return R?s.value?d/2-c.value/2:A:l&&typeof l.paddingLeftRef.value=="number"?k/2+l.paddingLeftRef.value:a&&typeof a.paddingLeftRef.value=="number"?(N?k/2:k)+a.paddingLeftRef.value:0}),w=x(()=>{const{collapsedWidth:d,indent:k,rootIndent:y}=t,{value:R}=c,{root:N}=e;return m.value||!N||!s.value?se:(y===void 0?k:y)+R+se-(d+R)/2});return{dropdownPlacement:h,activeIconSize:f,maxIconSize:c,paddingLeft:_,iconMarginRight:w,NMenu:o,NSubmenu:a,NMenuOptionGroup:l}}const Se={internalKey:{type:[String,Number],required:!0},root:Boolean,isGroup:Boolean,level:{type:Number,required:!0},title:[String,Function],extra:[String,Function]},rt=L({name:"MenuDivider",setup(){const e=D(le),{mergedClsPrefixRef:o,isHorizontalRef:t}=e;return()=>t.value?null:i("div",{class:`${o.value}-menu-divider`})}}),Ge=Object.assign(Object.assign({},Se),{tmNode:{type:Object,required:!0},disabled:Boolean,icon:Function,onClick:Function}),nt=xe(Ge),lt=L({name:"MenuOption",props:Ge,setup(e){const o=Ie(e),{NSubmenu:t,NMenu:s,NMenuOptionGroup:a}=o,{props:l,mergedClsPrefixRef:m,mergedCollapsedRef:h}=s,c=t?t.mergedDisabledRef:a?a.mergedDisabledRef:{value:!1},f=x(()=>c.value||e.disabled);function _(d){const{onClick:k}=e;k&&k(d)}function w(d){f.value||(s.doSelect(e.internalKey,e.tmNode.rawNode),_(d))}return{mergedClsPrefix:m,dropdownPlacement:o.dropdownPlacement,paddingLeft:o.paddingLeft,iconMarginRight:o.iconMarginRight,maxIconSize:o.maxIconSize,activeIconSize:o.activeIconSize,mergedTheme:s.mergedThemeRef,menuProps:l,dropdownEnabled:fe(()=>e.root&&h.value&&l.mode!=="horizontal"&&!f.value),selected:fe(()=>s.mergedValueRef.value===e.internalKey),mergedDisabled:f,handleClick:w}},render(){const{mergedClsPrefix:e,mergedTheme:o,tmNode:t,menuProps:{renderLabel:s,nodeProps:a}}=this,l=a==null?void 0:a(t.rawNode);return i("div",Object.assign({},l,{role:"menuitem",class:[`${e}-menu-item`,l==null?void 0:l.class]}),i(To,{theme:o.peers.Tooltip,themeOverrides:o.peerOverrides.Tooltip,trigger:"hover",placement:this.dropdownPlacement,disabled:!this.dropdownEnabled||this.title===void 0,internalExtraClass:["menu-tooltip"]},{default:()=>s?s(t.rawNode):Z(this.title),trigger:()=>i(Ue,{tmNode:t,clsPrefix:e,paddingLeft:this.paddingLeft,iconMarginRight:this.iconMarginRight,maxIconSize:this.maxIconSize,activeIconSize:this.activeIconSize,selected:this.selected,title:this.title,extra:this.extra,disabled:this.mergedDisabled,icon:this.icon,onClick:this.handleClick})}))}}),qe=Object.assign(Object.assign({},Se),{tmNode:{type:Object,required:!0},tmNodes:{type:Array,required:!0}}),it=xe(qe),at=L({name:"MenuOptionGroup",props:qe,setup(e){const o=Ie(e),{NSubmenu:t}=o,s=x(()=>t!=null&&t.mergedDisabledRef.value?!0:e.tmNode.disabled);W(ze,{paddingLeftRef:o.paddingLeft,mergedDisabledRef:s});const{mergedClsPrefixRef:a,props:l}=D(le);return function(){const{value:m}=a,h=o.paddingLeft.value,{nodeProps:c}=l,f=c==null?void 0:c(e.tmNode.rawNode);return i("div",{class:`${m}-menu-item-group`,role:"group"},i("div",Object.assign({},f,{class:[`${m}-menu-item-group-title`,f==null?void 0:f.class],style:[(f==null?void 0:f.style)||"",h!==void 0?`padding-left: ${h}px;`:""]}),Z(e.title),e.extra?i($e,null," ",Z(e.extra)):null),i("div",null,e.tmNodes.map(_=>we(_,l))))}}});function ge(e){return e.type==="divider"||e.type==="render"}function st(e){return e.type==="divider"}function we(e,o){const{rawNode:t}=e,{show:s}=t;if(s===!1)return null;if(ge(t))return st(t)?i(rt,Object.assign({key:e.key},t.props)):null;const{labelField:a}=o,{key:l,level:m,isGroup:h}=e,c=Object.assign(Object.assign({},t),{title:t.title||t[a],extra:t.titleExtra||t.extra,key:l,internalKey:l,level:m,root:m===0,isGroup:h});return e.children?e.isGroup?i(at,de(c,it,{tmNode:e,tmNodes:e.children,key:l})):i(be,de(c,ct,{key:l,rawNodes:t[o.childrenField],tmNodes:e.children,tmNode:e})):i(lt,de(c,nt,{key:l,tmNode:e}))}const Ye=Object.assign(Object.assign({},Se),{rawNodes:{type:Array,default:()=>[]},tmNodes:{type:Array,default:()=>[]},tmNode:{type:Object,required:!0},disabled:Boolean,icon:Function,onClick:Function,domId:String,virtualChildActive:{type:Boolean,default:void 0},isEllipsisPlaceholder:Boolean}),ct=xe(Ye),be=L({name:"Submenu",props:Ye,setup(e){const o=Ie(e),{NMenu:t,NSubmenu:s}=o,{props:a,mergedCollapsedRef:l,mergedThemeRef:m}=t,h=x(()=>{const{disabled:d}=e;return s!=null&&s.mergedDisabledRef.value||a.disabled?!0:d}),c=F(!1);W(De,{paddingLeftRef:o.paddingLeft,mergedDisabledRef:h}),W(ze,null);function f(){const{onClick:d}=e;d&&d()}function _(){h.value||(l.value||t.toggleExpand(e.internalKey),f())}function w(d){c.value=d}return{menuProps:a,mergedTheme:m,doSelect:t.doSelect,inverted:t.invertedRef,isHorizontal:t.isHorizontalRef,mergedClsPrefix:t.mergedClsPrefixRef,maxIconSize:o.maxIconSize,activeIconSize:o.activeIconSize,iconMarginRight:o.iconMarginRight,dropdownPlacement:o.dropdownPlacement,dropdownShow:c,paddingLeft:o.paddingLeft,mergedDisabled:h,mergedValue:t.mergedValueRef,childActive:fe(()=>{var d;return(d=e.virtualChildActive)!==null&&d!==void 0?d:t.activePathRef.value.includes(e.internalKey)}),collapsed:x(()=>a.mode==="horizontal"?!1:l.value?!0:!t.mergedExpandedKeysRef.value.includes(e.internalKey)),dropdownEnabled:x(()=>!h.value&&(a.mode==="horizontal"||l.value)),handlePopoverShowChange:w,handleClick:_}},render(){var e;const{mergedClsPrefix:o,menuProps:{renderIcon:t,renderLabel:s}}=this,a=()=>{const{isHorizontal:m,paddingLeft:h,collapsed:c,mergedDisabled:f,maxIconSize:_,activeIconSize:w,title:d,childActive:k,icon:y,handleClick:R,menuProps:{nodeProps:N},dropdownShow:A,iconMarginRight:Q,tmNode:U,mergedClsPrefix:M,isEllipsisPlaceholder:B,extra:z}=this,I=N==null?void 0:N(U.rawNode);return i("div",Object.assign({},I,{class:[`${M}-menu-item`,I==null?void 0:I.class],role:"menuitem"}),i(Ue,{tmNode:U,paddingLeft:h,collapsed:c,disabled:f,iconMarginRight:Q,maxIconSize:_,activeIconSize:w,title:d,extra:z,showArrow:!m,childActive:k,clsPrefix:M,icon:y,hover:A,onClick:R,isEllipsisPlaceholder:B}))},l=()=>i(fo,null,{default:()=>{const{tmNodes:m,collapsed:h}=this;return h?null:i("div",{class:`${o}-submenu-children`,role:"menu"},m.map(c=>we(c,this.menuProps)))}});return this.root?i(Fe,Object.assign({size:"large",trigger:"hover"},(e=this.menuProps)===null||e===void 0?void 0:e.dropdownProps,{themeOverrides:this.mergedTheme.peerOverrides.Dropdown,theme:this.mergedTheme.peers.Dropdown,builtinThemeOverrides:{fontSizeLarge:"14px",optionIconSizeLarge:"18px"},value:this.mergedValue,disabled:!this.dropdownEnabled,placement:this.dropdownPlacement,keyField:this.menuProps.keyField,labelField:this.menuProps.labelField,childrenField:this.menuProps.childrenField,onUpdateShow:this.handlePopoverShowChange,options:this.rawNodes,onSelect:this.doSelect,inverted:this.inverted,renderIcon:t,renderLabel:s}),{default:()=>i("div",{class:`${o}-submenu`,role:"menu","aria-expanded":!this.collapsed,id:this.domId},a(),this.isHorizontal?null:l())}):i("div",{class:`${o}-submenu`,role:"menu","aria-expanded":!this.collapsed,id:this.domId},a(),l())}}),dt=Object.assign(Object.assign({},K.props),{options:{type:Array,default:()=>[]},collapsed:{type:Boolean,default:void 0},collapsedWidth:{type:Number,default:48},iconSize:{type:Number,default:20},collapsedIconSize:{type:Number,default:24},rootIndent:Number,indent:{type:Number,default:32},labelField:{type:String,default:"label"},keyField:{type:String,default:"key"},childrenField:{type:String,default:"children"},disabledField:{type:String,default:"disabled"},defaultExpandAll:Boolean,defaultExpandedKeys:Array,expandedKeys:Array,value:[String,Number],defaultValue:{type:[String,Number],default:null},mode:{type:String,default:"vertical"},watchProps:{type:Array,default:void 0},disabled:Boolean,show:{type:Boolean,default:!0},inverted:Boolean,"onUpdate:expandedKeys":[Function,Array],onUpdateExpandedKeys:[Function,Array],onUpdateValue:[Function,Array],"onUpdate:value":[Function,Array],expandIcon:Function,renderIcon:Function,renderLabel:Function,renderExtra:Function,dropdownProps:Object,accordion:Boolean,nodeProps:Function,dropdownPlacement:{type:String,default:"bottom"},responsive:Boolean,items:Array,onOpenNamesChange:[Function,Array],onSelect:[Function,Array],onExpandedNamesChange:[Function,Array],expandedNames:Array,defaultExpandedNames:Array}),ut=L({name:"Menu",inheritAttrs:!1,props:dt,setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:t}=re(e),s=K("Menu","-menu",tt,xo,e,o),a=D(je,null),l=x(()=>{var p;const{collapsed:C}=e;if(C!==void 0)return C;if(a){const{collapseModeRef:r,collapsedRef:g}=a;if(r.value==="width")return(p=g.value)!==null&&p!==void 0?p:!1}return!1}),m=x(()=>{const{keyField:p,childrenField:C,disabledField:r}=e;return he(e.items||e.options,{getIgnored(g){return ge(g)},getChildren(g){return g[C]},getDisabled(g){return g[r]},getKey(g){var T;return(T=g[p])!==null&&T!==void 0?T:g.name}})}),h=x(()=>new Set(m.value.treeNodes.map(p=>p.key))),{watchProps:c}=e,f=F(null);c!=null&&c.includes("defaultValue")?Pe(()=>{f.value=e.defaultValue}):f.value=e.defaultValue;const _=te(e,"value"),w=pe(_,f),d=F([]),k=()=>{d.value=e.defaultExpandAll?m.value.getNonLeafKeys():e.defaultExpandedNames||e.defaultExpandedKeys||m.value.getPath(w.value,{includeSelf:!1}).keyPath};c!=null&&c.includes("defaultExpandedKeys")?Pe(k):k();const y=Bo(e,["expandedNames","expandedKeys"]),R=pe(y,d),N=x(()=>m.value.treeNodes),A=x(()=>m.value.getPath(w.value).keyPath);W(le,{props:e,mergedCollapsedRef:l,mergedThemeRef:s,mergedValueRef:w,mergedExpandedKeysRef:R,activePathRef:A,mergedClsPrefixRef:o,isHorizontalRef:x(()=>e.mode==="horizontal"),invertedRef:te(e,"inverted"),doSelect:Q,toggleExpand:M});function Q(p,C){const{"onUpdate:value":r,onUpdateValue:g,onSelect:T}=e;g&&j(g,p,C),r&&j(r,p,C),T&&j(T,p,C),f.value=p}function U(p){const{"onUpdate:expandedKeys":C,onUpdateExpandedKeys:r,onExpandedNamesChange:g,onOpenNamesChange:T}=e;C&&j(C,p),r&&j(r,p),g&&j(g,p),T&&j(T,p),d.value=p}function M(p){const C=Array.from(R.value),r=C.findIndex(g=>g===p);if(~r)C.splice(r,1);else{if(e.accordion&&h.value.has(p)){const g=C.findIndex(T=>h.value.has(T));g>-1&&C.splice(g,1)}C.push(p)}U(C)}const B=p=>{const C=m.value.getPath(p??w.value,{includeSelf:!1}).keyPath;if(!C.length)return;const r=Array.from(R.value),g=new Set([...r,...C]);e.accordion&&h.value.forEach(T=>{g.has(T)&&!C.includes(T)&&g.delete(T)}),U(Array.from(g))},z=x(()=>{const{inverted:p}=e,{common:{cubicBezierEaseInOut:C},self:r}=s.value,{borderRadius:g,borderColorHorizontal:T,fontSize:to,itemHeight:ro,dividerColor:no}=r,n={"--n-divider-color":no,"--n-bezier":C,"--n-font-size":to,"--n-border-color-horizontal":T,"--n-border-radius":g,"--n-item-height":ro};return p?(n["--n-group-text-color"]=r.groupTextColorInverted,n["--n-color"]=r.colorInverted,n["--n-item-text-color"]=r.itemTextColorInverted,n["--n-item-text-color-hover"]=r.itemTextColorHoverInverted,n["--n-item-text-color-active"]=r.itemTextColorActiveInverted,n["--n-item-text-color-child-active"]=r.itemTextColorChildActiveInverted,n["--n-item-text-color-child-active-hover"]=r.itemTextColorChildActiveInverted,n["--n-item-text-color-active-hover"]=r.itemTextColorActiveHoverInverted,n["--n-item-icon-color"]=r.itemIconColorInverted,n["--n-item-icon-color-hover"]=r.itemIconColorHoverInverted,n["--n-item-icon-color-active"]=r.itemIconColorActiveInverted,n["--n-item-icon-color-active-hover"]=r.itemIconColorActiveHoverInverted,n["--n-item-icon-color-child-active"]=r.itemIconColorChildActiveInverted,n["--n-item-icon-color-child-active-hover"]=r.itemIconColorChildActiveHoverInverted,n["--n-item-icon-color-collapsed"]=r.itemIconColorCollapsedInverted,n["--n-item-text-color-horizontal"]=r.itemTextColorHorizontalInverted,n["--n-item-text-color-hover-horizontal"]=r.itemTextColorHoverHorizontalInverted,n["--n-item-text-color-active-horizontal"]=r.itemTextColorActiveHorizontalInverted,n["--n-item-text-color-child-active-horizontal"]=r.itemTextColorChildActiveHorizontalInverted,n["--n-item-text-color-child-active-hover-horizontal"]=r.itemTextColorChildActiveHoverHorizontalInverted,n["--n-item-text-color-active-hover-horizontal"]=r.itemTextColorActiveHoverHorizontalInverted,n["--n-item-icon-color-horizontal"]=r.itemIconColorHorizontalInverted,n["--n-item-icon-color-hover-horizontal"]=r.itemIconColorHoverHorizontalInverted,n["--n-item-icon-color-active-horizontal"]=r.itemIconColorActiveHorizontalInverted,n["--n-item-icon-color-active-hover-horizontal"]=r.itemIconColorActiveHoverHorizontalInverted,n["--n-item-icon-color-child-active-horizontal"]=r.itemIconColorChildActiveHorizontalInverted,n["--n-item-icon-color-child-active-hover-horizontal"]=r.itemIconColorChildActiveHoverHorizontalInverted,n["--n-arrow-color"]=r.arrowColorInverted,n["--n-arrow-color-hover"]=r.arrowColorHoverInverted,n["--n-arrow-color-active"]=r.arrowColorActiveInverted,n["--n-arrow-color-active-hover"]=r.arrowColorActiveHoverInverted,n["--n-arrow-color-child-active"]=r.arrowColorChildActiveInverted,n["--n-arrow-color-child-active-hover"]=r.arrowColorChildActiveHoverInverted,n["--n-item-color-hover"]=r.itemColorHoverInverted,n["--n-item-color-active"]=r.itemColorActiveInverted,n["--n-item-color-active-hover"]=r.itemColorActiveHoverInverted,n["--n-item-color-active-collapsed"]=r.itemColorActiveCollapsedInverted):(n["--n-group-text-color"]=r.groupTextColor,n["--n-color"]=r.color,n["--n-item-text-color"]=r.itemTextColor,n["--n-item-text-color-hover"]=r.itemTextColorHover,n["--n-item-text-color-active"]=r.itemTextColorActive,n["--n-item-text-color-child-active"]=r.itemTextColorChildActive,n["--n-item-text-color-child-active-hover"]=r.itemTextColorChildActiveHover,n["--n-item-text-color-active-hover"]=r.itemTextColorActiveHover,n["--n-item-icon-color"]=r.itemIconColor,n["--n-item-icon-color-hover"]=r.itemIconColorHover,n["--n-item-icon-color-active"]=r.itemIconColorActive,n["--n-item-icon-color-active-hover"]=r.itemIconColorActiveHover,n["--n-item-icon-color-child-active"]=r.itemIconColorChildActive,n["--n-item-icon-color-child-active-hover"]=r.itemIconColorChildActiveHover,n["--n-item-icon-color-collapsed"]=r.itemIconColorCollapsed,n["--n-item-text-color-horizontal"]=r.itemTextColorHorizontal,n["--n-item-text-color-hover-horizontal"]=r.itemTextColorHoverHorizontal,n["--n-item-text-color-active-horizontal"]=r.itemTextColorActiveHorizontal,n["--n-item-text-color-child-active-horizontal"]=r.itemTextColorChildActiveHorizontal,n["--n-item-text-color-child-active-hover-horizontal"]=r.itemTextColorChildActiveHoverHorizontal,n["--n-item-text-color-active-hover-horizontal"]=r.itemTextColorActiveHoverHorizontal,n["--n-item-icon-color-horizontal"]=r.itemIconColorHorizontal,n["--n-item-icon-color-hover-horizontal"]=r.itemIconColorHoverHorizontal,n["--n-item-icon-color-active-horizontal"]=r.itemIconColorActiveHorizontal,n["--n-item-icon-color-active-hover-horizontal"]=r.itemIconColorActiveHoverHorizontal,n["--n-item-icon-color-child-active-horizontal"]=r.itemIconColorChildActiveHorizontal,n["--n-item-icon-color-child-active-hover-horizontal"]=r.itemIconColorChildActiveHoverHorizontal,n["--n-arrow-color"]=r.arrowColor,n["--n-arrow-color-hover"]=r.arrowColorHover,n["--n-arrow-color-active"]=r.arrowColorActive,n["--n-arrow-color-active-hover"]=r.arrowColorActiveHover,n["--n-arrow-color-child-active"]=r.arrowColorChildActive,n["--n-arrow-color-child-active-hover"]=r.arrowColorChildActiveHover,n["--n-item-color-hover"]=r.itemColorHover,n["--n-item-color-active"]=r.itemColorActive,n["--n-item-color-active-hover"]=r.itemColorActiveHover,n["--n-item-color-active-collapsed"]=r.itemColorActiveCollapsed),n}),I=t?ne("menu",x(()=>e.inverted?"a":"b"),z,e):void 0,G=go(),V=F(null),ce=F(null);let $=!0;const Re=()=>{var p;$?$=!1:(p=V.value)===null||p===void 0||p.sync({showAllItemsBeforeCalculate:!0})};function We(){return document.getElementById(G)}const ie=F(-1);function Xe(p){ie.value=e.options.length-p}function Ze(p){p||(ie.value=-1)}const Je=x(()=>{const p=ie.value;return{children:p===-1?[]:e.options.slice(p)}}),Qe=x(()=>{const{childrenField:p,disabledField:C,keyField:r}=e;return he([Je.value],{getIgnored(g){return ge(g)},getChildren(g){return g[p]},getDisabled(g){return g[C]},getKey(g){var T;return(T=g[r])!==null&&T!==void 0?T:g.name}})}),eo=x(()=>he([{}]).treeNodes[0]);function oo(){var p;if(ie.value===-1)return i(be,{root:!0,level:0,key:"__ellpisisGroupPlaceholder__",internalKey:"__ellpisisGroupPlaceholder__",title:"···",tmNode:eo.value,domId:G,isEllipsisPlaceholder:!0});const C=Qe.value.treeNodes[0],r=A.value,g=!!(!((p=C.children)===null||p===void 0)&&p.some(T=>r.includes(T.key)));return i(be,{level:0,root:!0,key:"__ellpisisGroup__",internalKey:"__ellpisisGroup__",title:"···",virtualChildActive:g,tmNode:C,domId:G,rawNodes:C.rawNode.children||[],tmNodes:C.children||[],isEllipsisPlaceholder:!0})}return{mergedClsPrefix:o,controlledExpandedKeys:y,uncontrolledExpanededKeys:d,mergedExpandedKeys:R,uncontrolledValue:f,mergedValue:w,activePath:A,tmNodes:N,mergedTheme:s,mergedCollapsed:l,cssVars:t?void 0:z,themeClass:I==null?void 0:I.themeClass,overflowRef:V,counterRef:ce,updateCounter:()=>{},onResize:Re,onUpdateOverflow:Ze,onUpdateCount:Xe,renderCounter:oo,getCounter:We,onRender:I==null?void 0:I.onRender,showOption:B,deriveResponsiveState:Re}},render(){const{mergedClsPrefix:e,mode:o,themeClass:t,onRender:s}=this;s==null||s();const a=()=>this.tmNodes.map(c=>we(c,this.$props)),m=o==="horizontal"&&this.responsive,h=()=>i("div",bo(this.$attrs,{role:o==="horizontal"?"menubar":"menu",class:[`${e}-menu`,t,`${e}-menu--${o}`,m&&`${e}-menu--responsive`,this.mergedCollapsed&&`${e}-menu--collapsed`],style:this.cssVars}),m?i(Ao,{ref:"overflowRef",onUpdateOverflow:this.onUpdateOverflow,getCounter:this.getCounter,onUpdateCount:this.onUpdateCount,updateCounter:this.updateCounter,style:{width:"100%",display:"flex",overflow:"hidden"}},{default:a,counter:this.renderCounter}):a());return m?i(po,{onResize:this.onResize},{default:h}):h()}}),vt={class:"h-14 flex items-center justify-center px-4 border-b border-border"},mt={key:1,class:"flex-1"},ht=L({__name:"AppLayout",setup(e){const{t:o}=Co(),t=No(),s=Po(),a=yo(),l=zo(),m=[{label:()=>i(O,{to:"/"},{default:()=>o("nav.dashboard")}),key:"dashboard"},{label:()=>i(O,{to:"/users"},{default:()=>o("nav.users")}),key:"users"},{label:()=>i(O,{to:"/groups"},{default:()=>o("nav.groups")}),key:"groups"},{label:()=>i(O,{to:"/channels"},{default:()=>o("nav.channels")}),key:"channels"},{label:()=>i(O,{to:"/accounts"},{default:()=>o("nav.accounts")}),key:"accounts"},{label:()=>i(O,{to:"/models"},{default:()=>o("nav.models")}),key:"models"},{label:()=>i(O,{to:"/pricing"},{default:()=>o("nav.pricing")}),key:"pricing"},{label:()=>i(O,{to:"/tokens"},{default:()=>o("nav.tokens")}),key:"tokens"},{label:o("nav.logs"),key:"logs",children:[{label:()=>i(O,{to:"/logs/requests"},{default:()=>o("nav.logs_requests")}),key:"logs-requests"},{label:()=>i(O,{to:"/logs/errors"},{default:()=>o("nav.logs_errors")}),key:"logs-errors"},{label:()=>i(O,{to:"/logs/audit"},{default:()=>o("nav.logs_audit")}),key:"logs-audit"}]},{label:()=>i(O,{to:"/stats"},{default:()=>o("nav.stats")}),key:"stats"},{label:()=>i(O,{to:"/notices"},{default:()=>o("nav.notices")}),key:"notices"},{label:o("nav.payments"),key:"payments",children:[{label:()=>i(O,{to:"/payments/recharges"},{default:()=>o("nav.payments_recharges")}),key:"payments-recharges"},{label:()=>i(O,{to:"/payments/redeem"},{default:()=>o("nav.payments_redeem")}),key:"payments-redeem"}]},{label:()=>i(O,{to:"/settings"},{default:()=>o("nav.settings")}),key:"settings"}],h=x(()=>String(t.name??"dashboard")),c=[{label:o("nav.profile"),key:"profile"},{label:o("nav.logout"),key:"logout"}];async function f(w){w==="logout"?(await l.logout(),s.push({name:"login"})):w==="profile"&&s.push({name:"profile"})}const _=x(()=>{var d;const w=(d=t.meta)==null?void 0:d.breadcrumb;return w?w.map(k=>({label:k.startsWith(":")?t.params[k.slice(1)]:o(`nav.${k}`),key:k})):[]});return(w,d)=>{const k=ko("RouterView");return q(),X(S(Ae),{"has-sider":"",style:{height:"100vh"}},{default:H(()=>[E(S(ot),{collapsed:S(a).sidebarCollapsed,"collapse-mode":"width","collapsed-width":64,width:220,bordered:"","show-trigger":"bar","onUpdate:collapsed":S(a).toggleSidebar},{default:H(()=>[ue("div",vt,[S(a).sidebarCollapsed?(q(),X(S(Te),{key:1,strong:""},{default:H(()=>[...d[2]||(d[2]=[oe("P",-1)])]),_:1})):(q(),X(S(Te),{key:0,strong:"",class:"text-lg"},{default:H(()=>[...d[1]||(d[1]=[oe("proapi",-1)])]),_:1}))]),E(S(ut),{collapsed:S(a).sidebarCollapsed,"collapsed-width":64,"collapsed-icon-size":20,options:m,value:h.value,"default-expanded-keys":["logs","payments"]},null,8,["collapsed","value"])]),_:1},8,["collapsed","onUpdate:collapsed"]),E(S(Ae),null,{default:H(()=>[E(S(Xo),{bordered:"",class:"px-4 h-14 flex items-center gap-3"},{default:H(()=>[E(S(ae),{text:"",onClick:S(a).toggleSidebar},{default:H(()=>[E(S(_e),{size:"20"},{default:H(()=>[...d[3]||(d[3]=[ue("span",{class:"i-lucide-menu"},null,-1)])]),_:1})]),_:1},8,["onClick"]),_.value.length?(q(),X(S(Mo),{key:0,class:"flex-1"},{default:H(()=>[(q(!0),Ne($e,null,Io(_.value,y=>(q(),X(S(Vo),{key:y.key},{default:H(()=>[oe(ve(y.label),1)]),_:2},1024))),128))]),_:1})):(q(),Ne("div",mt)),E(S(Ho),{align:"center",size:12},{default:H(()=>[E(S(ae),{text:"",onClick:S(a).toggleTheme},{default:H(()=>[E(S(_e),{size:"20"},{default:H(()=>[ue("span",{class:So(S(a).resolvedTheme==="dark"?"i-lucide-sun":"i-lucide-moon")},null,2)]),_:1})]),_:1},8,["onClick"]),E(S(ae),{text:"",onClick:d[0]||(d[0]=y=>S(a).setLocale(S(a).locale==="zh"?"en":"zh")),size:"small"},{default:H(()=>[oe(ve(S(a).locale==="zh"?"EN":"中"),1)]),_:1}),E(S(Fe),{options:c,onSelect:f},{default:H(()=>[E(S(ae),{text:""},{default:H(()=>[E(S(Oo),{type:"primary",size:"small"},{default:H(()=>{var y;return[oe(ve(((y=S(l).user)==null?void 0:y.username)??"--"),1)]}),_:1})]),_:1})]),_:1})]),_:1})]),_:1}),E(S(qo),{class:"p-4 overflow-auto",style:{height:"calc(100vh - 56px)"}},{default:H(()=>[E(k,null,{default:H(({Component:y})=>[E(wo,{name:"fade",mode:"out-in"},{default:H(()=>[(q(),X(Ro(y)))]),_:2},1024)]),_:1})]),_:1})]),_:1})]),_:1})}}}),_t=Eo(ht,[["__scopeId","data-v-bd54b200"]]);export{_t as default};
