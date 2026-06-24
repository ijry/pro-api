import{y as I,G as _,A as v,D as d,ak as R,aS as h,b as P,cs as E,cG as D,ap as S,cH as F,a2 as G,cu as M,cI as T,bG as j,a8 as q,af as l,cN as m,cq as a,f as H,a5 as C,ck as x,cC as O,cB as U,bI as k,a6 as A,a7 as L,B as N,ad as $,bX as w}from"./index-BMyk45kF.js";import{N as K}from"./Alert-DZw_pwSa.js";import{N as W,a as B}from"./FormItem-YWbjbrYx.js";import{a as z}from"./Input-C2HQZJje.js";import"./get-D86kArqK.js";import"./use-locale-DWTMmuPp.js";const X=I("divider",`
 position: relative;
 display: flex;
 width: 100%;
 box-sizing: border-box;
 font-size: 16px;
 color: var(--n-text-color);
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
`,[_("vertical",`
 margin-top: 24px;
 margin-bottom: 24px;
 `,[_("no-title",`
 display: flex;
 align-items: center;
 `)]),v("title",`
 display: flex;
 align-items: center;
 margin-left: 12px;
 margin-right: 12px;
 white-space: nowrap;
 font-weight: var(--n-font-weight);
 `),d("title-position-left",[v("line",[d("left",{width:"28px"})])]),d("title-position-right",[v("line",[d("right",{width:"28px"})])]),d("dashed",[v("line",`
 background-color: #0000;
 height: 0px;
 width: 100%;
 border-style: dashed;
 border-width: 1px 0 0;
 `)]),d("vertical",`
 display: inline-block;
 height: 1em;
 margin: 0 8px;
 vertical-align: middle;
 width: 1px;
 `),v("line",`
 border: none;
 transition: background-color .3s var(--n-bezier), border-color .3s var(--n-bezier);
 height: 1px;
 width: 100%;
 margin: 0;
 `),_("dashed",[v("line",{backgroundColor:"var(--n-color)"})]),d("dashed",[v("line",{borderColor:"var(--n-color)"})]),d("vertical",{backgroundColor:"var(--n-color)"})]),Y=Object.assign(Object.assign({},D.props),{titlePlacement:{type:String,default:"center"},dashed:Boolean,vertical:Boolean}),J=R({name:"Divider",props:Y,setup(r){const{mergedClsPrefixRef:i,inlineThemeDisabled:o}=E(r),n=D("Divider","-divider",X,S,r,i),c=G(()=>{const{common:{cubicBezierEaseInOut:e},self:{color:s,textColor:b,fontWeight:y}}=n.value;return{"--n-bezier":e,"--n-color":s,"--n-text-color":b,"--n-font-weight":y}}),t=o?F("divider",void 0,c,r):void 0;return{mergedClsPrefix:i,cssVars:o?void 0:c,themeClass:t==null?void 0:t.themeClass,onRender:t==null?void 0:t.onRender}},render(){var r;const{$slots:i,titlePlacement:o,vertical:n,dashed:c,cssVars:t,mergedClsPrefix:e}=this;return(r=this.onRender)===null||r===void 0||r.call(this),h("div",{role:"separator",class:[`${e}-divider`,this.themeClass,{[`${e}-divider--vertical`]:n,[`${e}-divider--no-title`]:!i.default,[`${e}-divider--dashed`]:c,[`${e}-divider--title-position-${o}`]:i.default&&o}],style:t},n?null:h("div",{class:`${e}-divider__line ${e}-divider__line--left`}),!n&&i.default?h(P,null,h("div",{class:`${e}-divider__title`},this.$slots),h("div",{class:`${e}-divider__line ${e}-divider__line--right`})):null)}}),Q={class:"min-h-screen flex flex-col items-center justify-center p-4"},Z={class:"text-sm opacity-60"},ee={class:"mt-4 text-sm opacity-40"},oe=R({__name:"Login",setup(r){const{t:i}=M(),o=O(),n=U(),c=T(),t=w({identity:"",password:""}),e=w(!1),s=w("");j(async()=>{c.user&&o.replace(n.query.redirect||"/")});async function b(){var g,u,f;if(!t.value.identity||!t.value.password){s.value="请填写用户名和密码";return}e.value=!0,s.value="";try{await c.login(t.value.identity,t.value.password),o.replace(n.query.redirect||"/")}catch(V){const p=V;((g=p==null?void 0:p.response)==null?void 0:g.status)===403?s.value=i("login.errors.not_admin"):(f=(u=p==null?void 0:p.response)==null?void 0:u.data)!=null&&f.message?s.value=p.response.data.message:s.value=i("login.errors.wrong_password")}finally{e.value=!1}}function y(g){g.key==="Enter"&&b()}return(g,u)=>(k(),q("div",Q,[l(a(H),{class:"w-full max-w-md",title:a(i)("login.title")},{"header-extra":m(()=>[C("span",Z,x(a(i)("login.subtitle")),1)]),default:m(()=>[s.value?(k(),A(a(K),{key:0,type:"error",class:"mb-4",title:s.value},null,8,["title"])):L("",!0),l(a(W),{"label-placement":"top",onKeydown:y},{default:m(()=>[l(a(B),{label:a(i)("login.identity")},{default:m(()=>[l(a(z),{value:t.value.identity,"onUpdate:value":u[0]||(u[0]=f=>t.value.identity=f),placeholder:a(i)("login.identity_placeholder"),disabled:e.value},null,8,["value","placeholder","disabled"])]),_:1},8,["label"]),l(a(B),{label:a(i)("login.password")},{default:m(()=>[l(a(z),{value:t.value.password,"onUpdate:value":u[1]||(u[1]=f=>t.value.password=f),type:"password","show-password-on":"click",disabled:e.value},null,8,["value","disabled"])]),_:1},8,["label"]),l(a(N),{type:"primary",block:"",loading:e.value,onClick:b},{default:m(()=>[$(x(a(i)("login.submit")),1)]),_:1},8,["loading"])]),_:1}),l(a(J)),l(a(N),{ghost:"",block:"",tag:"a",href:"/api/auth/oauth/github/start?redirect=/admin/"},{default:m(()=>[$(" GitHub "+x(a(i)("login.oauth_login")),1)]),_:1})]),_:1},8,["title"]),C("p",ee,"pro-api admin © "+x(new Date().getFullYear()),1)]))}});export{oe as default};
