import{x as p,y as e,G as N,D as z,A as B,aZ as V,a_ as F,ak as E,aE as H,aS as n,b$ as K,cs as W,cG as L,cH as Z,a2 as _,ao as q,ab as j}from"./index-BMyk45kF.js";import{g as J}from"./get-slot-Bk_rJcZu.js";import{u as Q}from"./cssr-E7GQDMMK.js";function M(t,h="default",d=[]){const{children:s}=t;if(s!==null&&typeof s=="object"&&!Array.isArray(s)){const i=s[h];if(typeof i=="function")return i()}return d}const U=p([e("descriptions",{fontSize:"var(--n-font-size)"},[e("descriptions-separator",`
 display: inline-block;
 margin: 0 8px 0 2px;
 `),e("descriptions-table-wrapper",[e("descriptions-table",[e("descriptions-table-row",[e("descriptions-table-header",{padding:"var(--n-th-padding)"}),e("descriptions-table-content",{padding:"var(--n-td-padding)"})])])]),N("bordered",[e("descriptions-table-wrapper",[e("descriptions-table",[e("descriptions-table-row",[p("&:last-child",[e("descriptions-table-content",{paddingBottom:0})])])])])]),z("left-label-placement",[e("descriptions-table-content",[p("> *",{verticalAlign:"top"})])]),z("left-label-align",[p("th",{textAlign:"left"})]),z("center-label-align",[p("th",{textAlign:"center"})]),z("right-label-align",[p("th",{textAlign:"right"})]),z("bordered",[e("descriptions-table-wrapper",`
 border-radius: var(--n-border-radius);
 overflow: hidden;
 background: var(--n-merged-td-color);
 border: 1px solid var(--n-merged-border-color);
 `,[e("descriptions-table",[e("descriptions-table-row",[p("&:not(:last-child)",[e("descriptions-table-content",{borderBottom:"1px solid var(--n-merged-border-color)"}),e("descriptions-table-header",{borderBottom:"1px solid var(--n-merged-border-color)"})]),e("descriptions-table-header",`
 font-weight: 400;
 background-clip: padding-box;
 background-color: var(--n-merged-th-color);
 `,[p("&:not(:last-child)",{borderRight:"1px solid var(--n-merged-border-color)"})]),e("descriptions-table-content",[p("&:not(:last-child)",{borderRight:"1px solid var(--n-merged-border-color)"})])])])])]),e("descriptions-header",`
 font-weight: var(--n-th-font-weight);
 font-size: 18px;
 transition: color .3s var(--n-bezier);
 line-height: var(--n-line-height);
 margin-bottom: 16px;
 color: var(--n-title-text-color);
 `),e("descriptions-table-wrapper",`
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `,[e("descriptions-table",`
 width: 100%;
 border-collapse: separate;
 border-spacing: 0;
 box-sizing: border-box;
 `,[e("descriptions-table-row",`
 box-sizing: border-box;
 transition: border-color .3s var(--n-bezier);
 `,[e("descriptions-table-header",`
 font-weight: var(--n-th-font-weight);
 line-height: var(--n-line-height);
 display: table-cell;
 box-sizing: border-box;
 color: var(--n-th-text-color);
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `),e("descriptions-table-content",`
 vertical-align: top;
 line-height: var(--n-line-height);
 display: table-cell;
 box-sizing: border-box;
 color: var(--n-td-text-color);
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `,[B("content",`
 transition: color .3s var(--n-bezier);
 display: inline-block;
 color: var(--n-td-text-color);
 `)]),B("label",`
 font-weight: var(--n-th-font-weight);
 transition: color .3s var(--n-bezier);
 display: inline-block;
 margin-right: 14px;
 color: var(--n-th-text-color);
 `)])])])]),e("descriptions-table-wrapper",`
 --n-merged-th-color: var(--n-th-color);
 --n-merged-td-color: var(--n-td-color);
 --n-merged-border-color: var(--n-border-color);
 `),V(e("descriptions-table-wrapper",`
 --n-merged-th-color: var(--n-th-color-modal);
 --n-merged-td-color: var(--n-td-color-modal);
 --n-merged-border-color: var(--n-border-color-modal);
 `)),F(e("descriptions-table-wrapper",`
 --n-merged-th-color: var(--n-th-color-popover);
 --n-merged-td-color: var(--n-td-color-popover);
 --n-merged-border-color: var(--n-border-color-popover);
 `))]),G="DESCRIPTION_ITEM_FLAG";function X(t){return typeof t=="object"&&t&&!Array.isArray(t)?t.type&&t.type[G]:!1}const Y=Object.assign(Object.assign({},L.props),{title:String,column:{type:Number,default:3},columns:Number,labelPlacement:{type:String,default:"top"},labelAlign:{type:String,default:"left"},separator:{type:String,default:":"},size:String,bordered:Boolean,labelClass:String,labelStyle:[Object,String],contentClass:String,contentStyle:[Object,String]}),ne=E({name:"Descriptions",props:Y,slots:Object,setup(t){const{mergedClsPrefixRef:h,inlineThemeDisabled:d,mergedComponentPropsRef:s}=W(t),i=_(()=>{var l,a;return t.size||((a=(l=s==null?void 0:s.value)===null||l===void 0?void 0:l.Descriptions)===null||a===void 0?void 0:a.size)||"medium"}),g=L("Descriptions","-descriptions",U,q,t,h),$=_(()=>{const{bordered:l}=t,a=i.value,{common:{cubicBezierEaseInOut:A},self:{titleTextColor:r,thColor:P,thColorModal:v,thColorPopover:R,thTextColor:I,thFontWeight:k,tdTextColor:D,tdColor:o,tdColorModal:f,tdColorPopover:O,borderColor:c,borderColorModal:m,borderColorPopover:y,borderRadius:w,lineHeight:u,[j("fontSize",a)]:S,[j(l?"thPaddingBordered":"thPadding",a)]:x,[j(l?"tdPaddingBordered":"tdPadding",a)]:C}}=g.value;return{"--n-title-text-color":r,"--n-th-padding":x,"--n-td-padding":C,"--n-font-size":S,"--n-bezier":A,"--n-th-font-weight":k,"--n-line-height":u,"--n-th-text-color":I,"--n-td-text-color":D,"--n-th-color":P,"--n-th-color-modal":v,"--n-th-color-popover":R,"--n-td-color":o,"--n-td-color-modal":f,"--n-td-color-popover":O,"--n-border-radius":w,"--n-border-color":c,"--n-border-color-modal":m,"--n-border-color-popover":y}}),b=d?Z("descriptions",_(()=>{let l="";const{bordered:a}=t;return a&&(l+="a"),l+=i.value[0],l}),$,t):void 0;return{mergedClsPrefix:h,cssVars:d?void 0:$,themeClass:b==null?void 0:b.themeClass,onRender:b==null?void 0:b.onRender,compitableColumn:Q(t,["columns","column"]),inlineThemeDisabled:d,mergedSize:i}},render(){const t=this.$slots.default,h=t?H(t()):[];h.length;const{contentClass:d,labelClass:s,compitableColumn:i,labelPlacement:g,labelAlign:$,mergedSize:b,bordered:l,title:a,cssVars:A,mergedClsPrefix:r,separator:P,onRender:v}=this;v==null||v();const R=h.filter(o=>X(o)),I={span:0,row:[],secondRow:[],rows:[]},D=R.reduce((o,f,O)=>{const c=f.props||{},m=R.length-1===O,y=["label"in c?c.label:M(f,"label")],w=[M(f)],u=c.span||1,S=o.span;o.span+=u;const x=c.labelStyle||c["label-style"]||this.labelStyle,C=c.contentStyle||c["content-style"]||this.contentStyle;if(g==="left")l?o.row.push(n("th",{class:[`${r}-descriptions-table-header`,s],colspan:1,style:x},y),n("td",{class:[`${r}-descriptions-table-content`,d],colspan:m?(i-S)*2+1:u*2-1,style:C},w)):o.row.push(n("td",{class:`${r}-descriptions-table-content`,colspan:m?(i-S)*2:u*2},n("span",{class:[`${r}-descriptions-table-content__label`,s],style:x},[...y,P&&n("span",{class:`${r}-descriptions-separator`},P)]),n("span",{class:[`${r}-descriptions-table-content__content`,d],style:C},w)));else{const T=m?(i-S)*2:u*2;o.row.push(n("th",{class:[`${r}-descriptions-table-header`,s],colspan:T,style:x},y)),o.secondRow.push(n("td",{class:[`${r}-descriptions-table-content`,d],colspan:T,style:C},w))}return(o.span>=i||m)&&(o.span=0,o.row.length&&(o.rows.push(o.row),o.row=[]),g!=="left"&&o.secondRow.length&&(o.rows.push(o.secondRow),o.secondRow=[])),o},I).rows.map(o=>n("tr",{class:`${r}-descriptions-table-row`},o));return n("div",{style:A,class:[`${r}-descriptions`,this.themeClass,`${r}-descriptions--${g}-label-placement`,`${r}-descriptions--${$}-label-align`,`${r}-descriptions--${b}-size`,l&&`${r}-descriptions--bordered`]},a||this.$slots.header?n("div",{class:`${r}-descriptions-header`},a||J(this,"header")):null,n("div",{class:`${r}-descriptions-table-wrapper`},n("table",{class:`${r}-descriptions-table`},n("tbody",null,g==="top"&&n("tr",{class:`${r}-descriptions-table-row`,style:{visibility:"collapse"}},K(i*2,n("td",null))),D))))}}),ee={label:String,span:{type:Number,default:1},labelClass:String,labelStyle:[Object,String],contentClass:String,contentStyle:[Object,String]},le=E({name:"DescriptionsItem",[G]:!0,props:ee,slots:Object,render(){return null}});export{ne as N,le as a};
