import{ak as _,cO as L,aS as s,F as ye,m as Q,bn as Se,S as Z,cJ as Y,bX as E,cs as ee,cD as Ce,cM as $e,cL as ze,bE as xe,cw as Be,a2 as $,aX as te,ar as X,O as ke,bS as P,aq as Re,bN as Ee,bt as Oe,x as i,R as j,y as g,D as C,A as B,aD as Te,cQ as Fe,L as He,b9 as Me,cG as re,cv as De,cH as Ne,az as Ie,H as k,as as Pe,cl as V,N as _e,ch as je,a6 as We,cN as R,cq as O,af as D,b_ as Ae,B as q,ad as K,ck as Ue,bI as Le}from"./index-BMyk45kF.js";import{u as G,f as J}from"./get-D86kArqK.js";import{N as Xe}from"./FormItem-YWbjbrYx.js";import{N as Ye}from"./Space-D5OWEM5p.js";const Ve=_({name:"NDrawerContent",inheritAttrs:!1,props:{blockScroll:Boolean,show:{type:Boolean,default:void 0},displayDirective:{type:String,required:!0},placement:{type:String,required:!0},contentClass:String,contentStyle:[Object,String],nativeScrollbar:{type:Boolean,required:!0},scrollbarProps:Object,trapFocus:{type:Boolean,default:!0},autoFocus:{type:Boolean,default:!0},showMask:{type:[Boolean,String],required:!0},maxWidth:Number,maxHeight:Number,minWidth:Number,minHeight:Number,resizable:Boolean,onClickoutside:Function,onAfterLeave:Function,onAfterEnter:Function,onEsc:Function},setup(e){const t=E(!!e.show),r=E(null),c=te(X);let f=0,p="",u=null;const h=E(!1),l=E(!1),y=$(()=>e.placement==="top"||e.placement==="bottom"),{mergedClsPrefixRef:T,mergedRtlRef:F}=ee(e),H=Ce("Drawer",F,T),z=o,v=n=>{l.value=!0,f=y.value?n.clientY:n.clientX,p=document.body.style.cursor,document.body.style.cursor=y.value?"ns-resize":"ew-resize",document.body.addEventListener("mousemove",S),document.body.addEventListener("mouseleave",z),document.body.addEventListener("mouseup",o)},N=()=>{u!==null&&(window.clearTimeout(u),u=null),l.value?h.value=!0:u=window.setTimeout(()=>{h.value=!0},300)},W=()=>{u!==null&&(window.clearTimeout(u),u=null),h.value=!1},{doUpdateHeight:A,doUpdateWidth:U}=c,M=n=>{const{maxWidth:d}=e;if(d&&n>d)return d;const{minWidth:m}=e;return m&&n<m?m:n},I=n=>{const{maxHeight:d}=e;if(d&&n>d)return d;const{minHeight:m}=e;return m&&n<m?m:n};function S(n){var d,m;if(l.value)if(y.value){let w=((d=r.value)===null||d===void 0?void 0:d.offsetHeight)||0;const x=f-n.clientY;w+=e.placement==="bottom"?x:-x,w=I(w),A(w),f=n.clientY}else{let w=((m=r.value)===null||m===void 0?void 0:m.offsetWidth)||0;const x=f-n.clientX;w+=e.placement==="right"?x:-x,w=M(w),U(w),f=n.clientX}}function o(){l.value&&(f=0,l.value=!1,document.body.style.cursor=p,document.body.removeEventListener("mousemove",S),document.body.removeEventListener("mouseup",o),document.body.removeEventListener("mouseleave",z))}$e(()=>{e.show&&(t.value=!0)}),ze(()=>e.show,n=>{n||o()}),xe(()=>{o()});const a=$(()=>{const{show:n}=e,d=[[Y,n]];return e.showMask||d.push([ke,e.onClickoutside,void 0,{capture:!0}]),d});function b(){var n;t.value=!1,(n=e.onAfterLeave)===null||n===void 0||n.call(e)}return Be($(()=>e.blockScroll&&t.value)),P(Re,r),P(Ee,null),P(Oe,null),{bodyRef:r,rtlEnabled:H,mergedClsPrefix:c.mergedClsPrefixRef,isMounted:c.isMountedRef,mergedTheme:c.mergedThemeRef,displayed:t,transitionName:$(()=>({right:"slide-in-from-right-transition",left:"slide-in-from-left-transition",top:"slide-in-from-top-transition",bottom:"slide-in-from-bottom-transition"})[e.placement]),handleAfterLeave:b,bodyDirectives:a,handleMousedownResizeTrigger:v,handleMouseenterResizeTrigger:N,handleMouseleaveResizeTrigger:W,isDragging:l,isHoverOnResizeTrigger:h}},render(){const{$slots:e,mergedClsPrefix:t}=this;return this.displayDirective==="show"||this.displayed||this.show?L(s("div",{role:"none"},s(ye,{disabled:!this.showMask||!this.trapFocus,active:this.show,autoFocus:this.autoFocus,onEsc:this.onEsc},{default:()=>s(Q,{name:this.transitionName,appear:this.isMounted,onAfterEnter:this.onAfterEnter,onAfterLeave:this.handleAfterLeave},{default:()=>L(s("div",Se(this.$attrs,{role:"dialog",ref:"bodyRef","aria-modal":"true",class:[`${t}-drawer`,this.rtlEnabled&&`${t}-drawer--rtl`,`${t}-drawer--${this.placement}-placement`,this.isDragging&&`${t}-drawer--unselectable`,this.nativeScrollbar&&`${t}-drawer--native-scrollbar`]}),[this.resizable?s("div",{class:[`${t}-drawer__resize-trigger`,(this.isDragging||this.isHoverOnResizeTrigger)&&`${t}-drawer__resize-trigger--hover`],onMouseenter:this.handleMouseenterResizeTrigger,onMouseleave:this.handleMouseleaveResizeTrigger,onMousedown:this.handleMousedownResizeTrigger}):null,this.nativeScrollbar?s("div",{class:[`${t}-drawer-content-wrapper`,this.contentClass],style:this.contentStyle,role:"none"},e):s(Z,Object.assign({},this.scrollbarProps,{contentStyle:this.contentStyle,contentClass:[`${t}-drawer-content-wrapper`,this.contentClass],theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar}),e)]),this.bodyDirectives)})})),[[Y,this.displayDirective==="if"||this.displayed||this.show]]):null}}),{cubicBezierEaseIn:qe,cubicBezierEaseOut:Ke}=j;function Ge({duration:e="0.3s",leaveDuration:t="0.2s",name:r="slide-in-from-bottom"}={}){return[i(`&.${r}-transition-leave-active`,{transition:`transform ${t} ${qe}`}),i(`&.${r}-transition-enter-active`,{transition:`transform ${e} ${Ke}`}),i(`&.${r}-transition-enter-to`,{transform:"translateY(0)"}),i(`&.${r}-transition-enter-from`,{transform:"translateY(100%)"}),i(`&.${r}-transition-leave-from`,{transform:"translateY(0)"}),i(`&.${r}-transition-leave-to`,{transform:"translateY(100%)"})]}const{cubicBezierEaseIn:Je,cubicBezierEaseOut:Qe}=j;function Ze({duration:e="0.3s",leaveDuration:t="0.2s",name:r="slide-in-from-left"}={}){return[i(`&.${r}-transition-leave-active`,{transition:`transform ${t} ${Je}`}),i(`&.${r}-transition-enter-active`,{transition:`transform ${e} ${Qe}`}),i(`&.${r}-transition-enter-to`,{transform:"translateX(0)"}),i(`&.${r}-transition-enter-from`,{transform:"translateX(-100%)"}),i(`&.${r}-transition-leave-from`,{transform:"translateX(0)"}),i(`&.${r}-transition-leave-to`,{transform:"translateX(-100%)"})]}const{cubicBezierEaseIn:et,cubicBezierEaseOut:tt}=j;function rt({duration:e="0.3s",leaveDuration:t="0.2s",name:r="slide-in-from-right"}={}){return[i(`&.${r}-transition-leave-active`,{transition:`transform ${t} ${et}`}),i(`&.${r}-transition-enter-active`,{transition:`transform ${e} ${tt}`}),i(`&.${r}-transition-enter-to`,{transform:"translateX(0)"}),i(`&.${r}-transition-enter-from`,{transform:"translateX(100%)"}),i(`&.${r}-transition-leave-from`,{transform:"translateX(0)"}),i(`&.${r}-transition-leave-to`,{transform:"translateX(100%)"})]}const{cubicBezierEaseIn:ot,cubicBezierEaseOut:nt}=j;function it({duration:e="0.3s",leaveDuration:t="0.2s",name:r="slide-in-from-top"}={}){return[i(`&.${r}-transition-leave-active`,{transition:`transform ${t} ${ot}`}),i(`&.${r}-transition-enter-active`,{transition:`transform ${e} ${nt}`}),i(`&.${r}-transition-enter-to`,{transform:"translateY(0)"}),i(`&.${r}-transition-enter-from`,{transform:"translateY(-100%)"}),i(`&.${r}-transition-leave-from`,{transform:"translateY(0)"}),i(`&.${r}-transition-leave-to`,{transform:"translateY(-100%)"})]}const st=i([g("drawer",`
 word-break: break-word;
 line-height: var(--n-line-height);
 position: absolute;
 pointer-events: all;
 box-shadow: var(--n-box-shadow);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 background-color: var(--n-color);
 color: var(--n-text-color);
 box-sizing: border-box;
 `,[rt(),Ze(),it(),Ge(),C("unselectable",`
 user-select: none; 
 -webkit-user-select: none;
 `),C("native-scrollbar",[g("drawer-content-wrapper",`
 overflow: auto;
 height: 100%;
 `)]),B("resize-trigger",`
 position: absolute;
 background-color: #0000;
 transition: background-color .3s var(--n-bezier);
 `,[C("hover",`
 background-color: var(--n-resize-trigger-color-hover);
 `)]),g("drawer-content-wrapper",`
 box-sizing: border-box;
 `),g("drawer-content",`
 height: 100%;
 display: flex;
 flex-direction: column;
 `,[C("native-scrollbar",[g("drawer-body-content-wrapper",`
 height: 100%;
 overflow: auto;
 `)]),g("drawer-body",`
 flex: 1 0 0;
 overflow: hidden;
 `),g("drawer-body-content-wrapper",`
 box-sizing: border-box;
 padding: var(--n-body-padding);
 `),g("drawer-header",`
 font-weight: var(--n-title-font-weight);
 line-height: 1;
 font-size: var(--n-title-font-size);
 color: var(--n-title-text-color);
 padding: var(--n-header-padding);
 transition: border .3s var(--n-bezier);
 border-bottom: 1px solid var(--n-divider-color);
 border-bottom: var(--n-header-border-bottom);
 display: flex;
 justify-content: space-between;
 align-items: center;
 `,[B("main",`
 flex: 1;
 `),B("close",`
 margin-left: 6px;
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `)]),g("drawer-footer",`
 display: flex;
 justify-content: flex-end;
 border-top: var(--n-footer-border-top);
 transition: border .3s var(--n-bezier);
 padding: var(--n-footer-padding);
 `)]),C("right-placement",`
 top: 0;
 bottom: 0;
 right: 0;
 border-top-left-radius: var(--n-border-radius);
 border-bottom-left-radius: var(--n-border-radius);
 `,[B("resize-trigger",`
 width: 3px;
 height: 100%;
 top: 0;
 left: 0;
 transform: translateX(-1.5px);
 cursor: ew-resize;
 `)]),C("left-placement",`
 top: 0;
 bottom: 0;
 left: 0;
 border-top-right-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `,[B("resize-trigger",`
 width: 3px;
 height: 100%;
 top: 0;
 right: 0;
 transform: translateX(1.5px);
 cursor: ew-resize;
 `)]),C("top-placement",`
 top: 0;
 left: 0;
 right: 0;
 border-bottom-left-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `,[B("resize-trigger",`
 width: 100%;
 height: 3px;
 bottom: 0;
 left: 0;
 transform: translateY(1.5px);
 cursor: ns-resize;
 `)]),C("bottom-placement",`
 left: 0;
 bottom: 0;
 right: 0;
 border-top-left-radius: var(--n-border-radius);
 border-top-right-radius: var(--n-border-radius);
 `,[B("resize-trigger",`
 width: 100%;
 height: 3px;
 top: 0;
 left: 0;
 transform: translateY(-1.5px);
 cursor: ns-resize;
 `)])]),i("body",[i(">",[g("drawer-container",`
 position: fixed;
 `)])]),g("drawer-container",`
 position: relative;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 pointer-events: none;
 `,[i("> *",`
 pointer-events: all;
 `)]),g("drawer-mask",`
 background-color: rgba(0, 0, 0, .3);
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `,[C("invisible",`
 background-color: rgba(0, 0, 0, 0)
 `),Te({enterDuration:"0.2s",leaveDuration:"0.2s",enterCubicBezier:"var(--n-bezier-in)",leaveCubicBezier:"var(--n-bezier-out)"})])]),at=Object.assign(Object.assign({},re.props),{show:Boolean,width:[Number,String],height:[Number,String],placement:{type:String,default:"right"},maskClosable:{type:Boolean,default:!0},showMask:{type:[Boolean,String],default:!0},to:[String,Object],displayDirective:{type:String,default:"if"},nativeScrollbar:{type:Boolean,default:!0},zIndex:Number,onMaskClick:Function,scrollbarProps:Object,contentClass:String,contentStyle:[Object,String],trapFocus:{type:Boolean,default:!0},onEsc:Function,autoFocus:{type:Boolean,default:!0},closeOnEsc:{type:Boolean,default:!0},blockScroll:{type:Boolean,default:!0},maxWidth:Number,maxHeight:Number,minWidth:Number,minHeight:Number,resizable:Boolean,defaultWidth:{type:[Number,String],default:251},defaultHeight:{type:[Number,String],default:251},onUpdateWidth:[Function,Array],onUpdateHeight:[Function,Array],"onUpdate:width":[Function,Array],"onUpdate:height":[Function,Array],"onUpdate:show":[Function,Array],onUpdateShow:[Function,Array],onAfterEnter:Function,onAfterLeave:Function,drawerStyle:[String,Object],drawerClass:String,target:null,onShow:Function,onHide:Function}),lt=_({name:"Drawer",inheritAttrs:!1,props:at,setup(e){const{mergedClsPrefixRef:t,namespaceRef:r,inlineThemeDisabled:c}=ee(e),f=Me(),p=re("Drawer","-drawer",st,Pe,e,t),u=E(e.defaultWidth),h=E(e.defaultHeight),l=G(V(e,"width"),u),y=G(V(e,"height"),h),T=$(()=>{const{placement:o}=e;return o==="top"||o==="bottom"?"":J(l.value)}),F=$(()=>{const{placement:o}=e;return o==="left"||o==="right"?"":J(y.value)}),H=o=>{const{onUpdateWidth:a,"onUpdate:width":b}=e;a&&k(a,o),b&&k(b,o),u.value=o},z=o=>{const{onUpdateHeight:a,"onUpdate:width":b}=e;a&&k(a,o),b&&k(b,o),h.value=o},v=$(()=>[{width:T.value,height:F.value},e.drawerStyle||""]);function N(o){const{onMaskClick:a,maskClosable:b}=e;b&&M(!1),a&&a(o)}function W(o){N(o)}const A=De();function U(o){var a;(a=e.onEsc)===null||a===void 0||a.call(e),e.show&&e.closeOnEsc&&Ie(o)&&(A.value||M(!1))}function M(o){const{onHide:a,onUpdateShow:b,"onUpdate:show":n}=e;b&&k(b,o),n&&k(n,o),a&&!o&&k(a,o)}P(X,{isMountedRef:f,mergedThemeRef:p,mergedClsPrefixRef:t,doUpdateShow:M,doUpdateHeight:z,doUpdateWidth:H});const I=$(()=>{const{common:{cubicBezierEaseInOut:o,cubicBezierEaseIn:a,cubicBezierEaseOut:b},self:{color:n,textColor:d,boxShadow:m,lineHeight:w,headerPadding:x,footerPadding:oe,borderRadius:ne,bodyPadding:ie,titleFontSize:se,titleTextColor:ae,titleFontWeight:le,headerBorderBottom:de,footerBorderTop:ce,closeIconColor:ue,closeIconColorHover:he,closeIconColorPressed:fe,closeColorHover:be,closeColorPressed:me,closeIconSize:ge,closeSize:ve,closeBorderRadius:we,resizableTriggerColorHover:pe}}=p.value;return{"--n-line-height":w,"--n-color":n,"--n-border-radius":ne,"--n-text-color":d,"--n-box-shadow":m,"--n-bezier":o,"--n-bezier-out":b,"--n-bezier-in":a,"--n-header-padding":x,"--n-body-padding":ie,"--n-footer-padding":oe,"--n-title-text-color":ae,"--n-title-font-size":se,"--n-title-font-weight":le,"--n-header-border-bottom":de,"--n-footer-border-top":ce,"--n-close-icon-color":ue,"--n-close-icon-color-hover":he,"--n-close-icon-color-pressed":fe,"--n-close-size":ve,"--n-close-color-hover":be,"--n-close-color-pressed":me,"--n-close-icon-size":ge,"--n-close-border-radius":we,"--n-resize-trigger-color-hover":pe}}),S=c?Ne("drawer",void 0,I,e):void 0;return{mergedClsPrefix:t,namespace:r,mergedBodyStyle:v,handleOutsideClick:W,handleMaskClick:N,handleEsc:U,mergedTheme:p,cssVars:c?void 0:I,themeClass:S==null?void 0:S.themeClass,onRender:S==null?void 0:S.onRender,isMounted:f}},render(){const{mergedClsPrefix:e}=this;return s(He,{to:this.to,show:this.show},{default:()=>{var t;return(t=this.onRender)===null||t===void 0||t.call(this),L(s("div",{class:[`${e}-drawer-container`,this.namespace,this.themeClass],style:this.cssVars,role:"none"},this.showMask?s(Q,{name:"fade-in-transition",appear:this.isMounted},{default:()=>this.show?s("div",{"aria-hidden":!0,class:[`${e}-drawer-mask`,this.showMask==="transparent"&&`${e}-drawer-mask--invisible`],onClick:this.handleMaskClick}):null}):null,s(Ve,Object.assign({},this.$attrs,{class:[this.drawerClass,this.$attrs.class],style:[this.mergedBodyStyle,this.$attrs.style],blockScroll:this.blockScroll,contentStyle:this.contentStyle,contentClass:this.contentClass,placement:this.placement,scrollbarProps:this.scrollbarProps,show:this.show,displayDirective:this.displayDirective,nativeScrollbar:this.nativeScrollbar,onAfterEnter:this.onAfterEnter,onAfterLeave:this.onAfterLeave,trapFocus:this.trapFocus,autoFocus:this.autoFocus,resizable:this.resizable,maxHeight:this.maxHeight,minHeight:this.minHeight,maxWidth:this.maxWidth,minWidth:this.minWidth,showMask:this.showMask,onEsc:this.handleEsc,onClickoutside:this.handleOutsideClick}),this.$slots)),[[Fe,{zIndex:this.zIndex,enabled:this.show}]])}})}}),dt={title:String,headerClass:String,headerStyle:[Object,String],footerClass:String,footerStyle:[Object,String],bodyClass:String,bodyStyle:[Object,String],bodyContentClass:String,bodyContentStyle:[Object,String],nativeScrollbar:{type:Boolean,default:!0},scrollbarProps:Object,closable:Boolean},ct=_({name:"DrawerContent",props:dt,slots:Object,setup(){const e=te(X,null);e||je("drawer-content","`n-drawer-content` must be placed inside `n-drawer`.");const{doUpdateShow:t}=e;function r(){t(!1)}return{handleCloseClick:r,mergedTheme:e.mergedThemeRef,mergedClsPrefix:e.mergedClsPrefixRef}},render(){const{title:e,mergedClsPrefix:t,nativeScrollbar:r,mergedTheme:c,bodyClass:f,bodyStyle:p,bodyContentClass:u,bodyContentStyle:h,headerClass:l,headerStyle:y,footerClass:T,footerStyle:F,scrollbarProps:H,closable:z,$slots:v}=this;return s("div",{role:"none",class:[`${t}-drawer-content`,r&&`${t}-drawer-content--native-scrollbar`]},v.header||e||z?s("div",{class:[`${t}-drawer-header`,l],style:y,role:"none"},s("div",{class:`${t}-drawer-header__main`,role:"heading","aria-level":"1"},v.header!==void 0?v.header():e),z&&s(_e,{onClick:this.handleCloseClick,clsPrefix:t,class:`${t}-drawer-header__close`,absolute:!0})):null,r?s("div",{class:[`${t}-drawer-body`,f],style:p,role:"none"},s("div",{class:[`${t}-drawer-body-content-wrapper`,u],style:h,role:"none"},v)):s(Z,Object.assign({themeOverrides:c.peerOverrides.Scrollbar,theme:c.peers.Scrollbar},H,{class:`${t}-drawer-body`,contentClass:[`${t}-drawer-body-content-wrapper`,u],contentStyle:h}),v),v.footer?s("div",{class:[`${t}-drawer-footer`,T],style:F,role:"none"},v.footer()):null)}}),mt=_({__name:"FormDrawer",props:{show:{type:Boolean},mode:{},title:{},loading:{type:Boolean},rules:{}},emits:["update:show","submit","cancel"],setup(e,{emit:t}){const r=e,c=t,f=E(null);function p(){c("update:show",!1),c("cancel")}async function u(){var h;try{await((h=f.value)==null?void 0:h.validate()),c("submit")}catch{}}return(h,l)=>(Le(),We(O(lt),{show:r.show,width:"480",placement:"right","onUpdate:show":l[0]||(l[0]=y=>c("update:show",y))},{default:R(()=>[D(O(ct),{title:r.title,closable:""},{footer:R(()=>[D(O(Ye),{justify:"end"},{default:R(()=>[D(O(q),{onClick:p},{default:R(()=>[...l[1]||(l[1]=[K("取消",-1)])]),_:1}),D(O(q),{type:"primary",loading:r.loading,onClick:u},{default:R(()=>[K(Ue(r.mode==="edit"?"保存":"创建"),1)]),_:1},8,["loading"])]),_:1})]),default:R(()=>[D(O(Xe),{ref_key:"formRef",ref:f,"label-placement":"top",rules:r.rules??{}},{default:R(()=>[Ae(h.$slots,"default")]),_:3},8,["rules"])]),_:3},8,["title"])]),_:3},8,["show"]))}});export{mt as _};
