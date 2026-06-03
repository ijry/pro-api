import{y as V,G as _,A as v,D as d,al as D,aU as h,b as E,cu as I,cI as R,ar as F,cJ as M,a3 as S,cw as T,cK as U,bI as j,a9 as K,ag as s,cP as m,cs as a,f as O,a6 as C,cm as x,cE as q,cD as A,bK as k,a7 as G,a8 as H,B as N,ae as $,bZ as w}from"./index-ClKeNBH9.js";import{N as L}from"./Alert-XMIW4hzt.js";import{N as J,a as B}from"./FormItem-jn2ynUsd.js";import{a as z}from"./Input-C00Z2Txa.js";import"./get-48VdzrSm.js";import"./use-locale-CGm--TNI.js";const W=V("divider",`
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
 `),_("dashed",[v("line",{backgroundColor:"var(--n-color)"})]),d("dashed",[v("line",{borderColor:"var(--n-color)"})]),d("vertical",{backgroundColor:"var(--n-color)"})]),Y=Object.assign(Object.assign({},R.props),{titlePlacement:{type:String,default:"center"},dashed:Boolean,vertical:Boolean}),Z=D({name:"Divider",props:Y,setup(r){const{mergedClsPrefixRef:i,inlineThemeDisabled:o}=I(r),n=R("Divider","-divider",W,F,r,i),c=S(()=>{const{common:{cubicBezierEaseInOut:e},self:{color:l,textColor:b,fontWeight:y}}=n.value;return{"--n-bezier":e,"--n-color":l,"--n-text-color":b,"--n-font-weight":y}}),t=o?M("divider",void 0,c,r):void 0;return{mergedClsPrefix:i,cssVars:o?void 0:c,themeClass:t==null?void 0:t.themeClass,onRender:t==null?void 0:t.onRender}},render(){var r;const{$slots:i,titlePlacement:o,vertical:n,dashed:c,cssVars:t,mergedClsPrefix:e}=this;return(r=this.onRender)===null||r===void 0||r.call(this),h("div",{role:"separator",class:[`${e}-divider`,this.themeClass,{[`${e}-divider--vertical`]:n,[`${e}-divider--no-title`]:!i.default,[`${e}-divider--dashed`]:c,[`${e}-divider--title-position-${o}`]:i.default&&o}],style:t},n?null:h("div",{class:`${e}-divider__line ${e}-divider__line--left`}),!n&&i.default?h(E,null,h("div",{class:`${e}-divider__title`},this.$slots),h("div",{class:`${e}-divider__line ${e}-divider__line--right`})):null)}}),Q={class:"min-h-screen flex flex-col items-center justify-center p-4"},X={class:"text-sm opacity-60"},ee={class:"mt-4 text-sm opacity-40"},oe=D({__name:"Login",setup(r){const{t:i}=T(),o=q(),n=A(),c=U(),t=w({identity:"",password:""}),e=w(!1),l=w("");j(async()=>{c.user&&o.replace(n.query.redirect||"/")});async function b(){var f,u,g;if(!t.value.identity||!t.value.password){l.value="请填写用户名和密码";return}e.value=!0,l.value="";try{await c.login(t.value.identity,t.value.password),o.replace(n.query.redirect||"/")}catch(P){const p=P;((f=p==null?void 0:p.response)==null?void 0:f.status)===403?l.value=i("login.errors.not_admin"):(g=(u=p==null?void 0:p.response)==null?void 0:u.data)!=null&&g.message?l.value=p.response.data.message:l.value=i("login.errors.wrong_password")}finally{e.value=!1}}function y(f){f.key==="Enter"&&b()}return(f,u)=>(k(),K("div",Q,[s(a(O),{class:"w-full max-w-md",title:a(i)("login.title")},{"header-extra":m(()=>[C("span",X,x(a(i)("login.subtitle")),1)]),default:m(()=>[l.value?(k(),G(a(L),{key:0,type:"error",class:"mb-4",title:l.value},null,8,["title"])):H("",!0),s(a(J),{"label-placement":"top",onKeydown:y},{default:m(()=>[s(a(B),{label:a(i)("login.identity")},{default:m(()=>[s(a(z),{value:t.value.identity,"onUpdate:value":u[0]||(u[0]=g=>t.value.identity=g),placeholder:a(i)("login.identity_placeholder"),disabled:e.value},null,8,["value","placeholder","disabled"])]),_:1},8,["label"]),s(a(B),{label:a(i)("login.password")},{default:m(()=>[s(a(z),{value:t.value.password,"onUpdate:value":u[1]||(u[1]=g=>t.value.password=g),type:"password","show-password-on":"click",disabled:e.value},null,8,["value","disabled"])]),_:1},8,["label"]),s(a(N),{type:"primary",block:"",loading:e.value,onClick:b},{default:m(()=>[$(x(a(i)("login.submit")),1)]),_:1},8,["loading"])]),_:1}),s(a(Z)),s(a(N),{ghost:"",block:"",tag:"a",href:"/api/auth/oauth/github/start?redirect=/admin/"},{default:m(()=>[$(" GitHub "+x(a(i)("login.oauth_login")),1)]),_:1})]),_:1},8,["title"]),C("p",ee,"proapi admin © "+x(new Date().getFullYear()),1)]))}});export{oe as default};
