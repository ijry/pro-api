import{l as M,j as y,H as $,r as B,f as l,F as x,i as r,b as e,B as a,D as s,x as f,s as n,R as H,y as k,d as N,A as L,h as F,o as I,c as q,G as O}from"./index-DNhoM-_O.js";import{_ as P}from"./WalletCard.vue_vue_type_script_setup_true_lang-tRNMfIQR.js";import{_ as C}from"./Skeleton.vue_vue_type_script_setup_true_lang-DLmQMa2_.js";import{l as S}from"./log-C5r2jM39.js";import{_ as E}from"./EmptyState.vue_vue_type_script_setup_true_lang-DQzPxrDZ.js";import{_ as R}from"./ClipboardButton.vue_vue_type_script_setup_true_lang-Diw-j5BZ.js";const U={get:d=>M(`/api/user/usage?range=${d}`)},z={class:"rounded-xl border border-border bg-bg-elevated backdrop-blur-md p-5 shadow-sm"},T={class:"text-xs text-fg-muted uppercase tracking-wide"},j={class:"text-2xl font-semibold text-fg mt-1"},D={class:"text-xs text-fg-muted mt-1"},A=y({__name:"UsageStatCard",props:{scope:{}},setup(d){const i=d,{t:o}=$(),c=f(null),u=f(!0);return B(async()=>{try{c.value=await U.get(i.scope)}finally{u.value=!1}}),(p,g)=>{var _,t;return n(),l("div",z,[u.value?(n(),l(x,{key:0},[r(C,{class:"h-6 w-24 mb-2"}),r(C,{class:"h-8 w-20"})],64)):(n(),l(x,{key:1},[e("p",T,a(d.scope==="today"?s(o)("home.usage.today"):s(o)("home.usage.month")),1),e("p",j," $ "+a((((_=c.value)==null?void 0:_.cost_usd)??0).toFixed(4)),1),e("p",D,a(s(o)("home.usage.requests",{n:(((t=c.value)==null?void 0:t.request_count)??0).toLocaleString()})),1)],64))])}}}),K={class:"rounded-xl border border-border bg-bg-elevated backdrop-blur-md p-5 shadow-sm"},V={class:"flex items-center justify-between mb-4"},G={class:"font-semibold text-fg"},W={key:0,class:"space-y-2"},Q={key:2,class:"space-y-1"},J={class:"text-fg flex-1 truncate"},X={class:"text-fg-muted text-xs"},Y={class:"text-fg-muted text-xs"},Z=y({__name:"RecentLogsCard",props:{limit:{default:5}},setup(d){const i=d,{t:o}=$(),c=f([]),u=f(!0);B(async()=>{try{const t=await S.list({page_size:i.limit});c.value=t.items}finally{u.value=!1}});function p(t){return t>=200&&t<300?"text-emerald-400":"text-rose-400"}function g(t){return t>=200&&t<300?"i-lucide-check-circle":"i-lucide-x-circle"}function _(t){return t>=1e3?`${(t/1e3).toFixed(1)}s`:`${t}ms`}return(t,h)=>{const w=L("router-link");return n(),l("div",K,[e("div",V,[e("h3",G,a(s(o)("home.logs.title")),1),r(w,{to:"/logs",class:"text-xs text-primary hover:underline"},{default:H(()=>[F(a(s(o)("home.logs.view_all")),1)]),_:1})]),u.value?(n(),l("div",W,[(n(),l(x,null,k(5,m=>r(C,{key:m,class:"h-10"})),64))])):c.value.length?(n(),l("div",Q,[(n(!0),l(x,null,k(c.value,m=>(n(),l("div",{key:m.id,class:"flex items-center gap-3 py-2 px-2 rounded-md hover:bg-bg transition-colors text-sm"},[e("span",{class:I([g(m.status),p(m.status),"w-4 h-4 shrink-0"])},null,2),e("span",J,a(m.model),1),e("span",X,"$ "+a(m.cost_usd.toFixed(4)),1),e("span",Y,a(_(m.latency_ms)),1)]))),128))])):(n(),N(E,{key:1,icon:"i-lucide-inbox",title:s(o)("home.logs.empty.title"),subtitle:s(o)("home.logs.empty.subtitle"),cta:s(o)("home.logs.empty.cta"),"cta-to":"/apikeys"},null,8,["title","subtitle","cta"]))])}}}),ee={class:"rounded-xl border border-border bg-bg-elevated backdrop-blur-md p-5 shadow-sm"},te={class:"font-semibold text-fg mb-4"},se={class:"flex gap-1 mb-3"},oe=["onClick"],ne={class:"relative rounded-lg bg-bg border border-border overflow-hidden"},ae={class:"absolute top-2 right-2"},ce={class:"text-xs text-fg-muted p-4 pr-10 overflow-x-auto leading-relaxed font-mono whitespace-pre"},le={class:"flex gap-2 mt-3"},b="https://api.proapi.io/v1",v="pa-xxx...",re=y({__name:"QuickStartCode",setup(d){const{t:i}=$(),o=["curl","python","node","go"],c=f("curl"),u={curl:`curl ${b}/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer ${v}" \\
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`,python:`from openai import OpenAI

client = OpenAI(
    base_url="${b}",
    api_key="${v}",
)

completion = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(completion.choices[0].message.content)`,node:`import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "${b}",
  apiKey: "${v}",
});

const completion = await client.chat.completions.create({
  model: "gpt-4o",
  messages: [{ role: "user", content: "Hello!" }],
});
console.log(completion.choices[0].message.content);`,go:`client := openai.NewClient(
    option.WithBaseURL("${b}"),
    option.WithAPIKey("${v}"),
)
completion, _ := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
    Model:    openai.F(openai.ChatModelGPT4o),
    Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Hello!"),
    }),
})
fmt.Println(completion.Choices[0].Message.Content)`},p=q(()=>u[c.value]);return(g,_)=>{const t=L("router-link");return n(),l("div",ee,[e("h3",te,a(s(i)("home.quickstart.title")),1),e("div",se,[(n(),l(x,null,k(o,h=>e("button",{key:h,onClick:w=>c.value=h,class:I(["px-3 h-7 rounded-md text-xs font-medium transition-colors",c.value===h?"bg-primary text-white":"text-fg-muted hover:text-fg hover:bg-bg"])},a(s(i)(`home.quickstart.tab.${h}`,h)),11,oe)),64))]),e("div",ne,[e("div",ae,[r(R,{text:p.value,"success-msg":s(i)("home.quickstart.copy"),size:"sm"},null,8,["text","success-msg"])]),e("pre",ce,a(p.value),1)]),e("div",le,[r(t,{to:"/apikeys",class:"inline-flex items-center gap-1 px-3 h-7 rounded-md border border-border text-xs text-fg hover:bg-bg transition-colors"},{default:H(()=>[_[0]||(_[0]=e("span",{class:"i-lucide-key-round w-3.5 h-3.5"},null,-1)),F(a(s(i)("home.quickstart.my_tokens")),1)]),_:1})])])}}}),ie={class:"space-y-6"},de={class:"text-2xl font-bold text-fg"},ue={class:"grid grid-cols-1 md:grid-cols-3 gap-4"},pe={class:"grid grid-cols-1 lg:grid-cols-2 gap-4"},be=y({__name:"index",setup(d){const{t:i}=$(),o=O();return(c,u)=>{var p,g;return n(),l("div",ie,[e("div",null,[e("h1",de,a(s(i)("home.greeting",{name:((p=s(o).user)==null?void 0:p.display_name)||((g=s(o).user)==null?void 0:g.email)||""})),1)]),e("div",ue,[r(P,{variant:"dashboard"}),r(A,{scope:"today"}),r(A,{scope:"month"})]),e("div",pe,[r(Z,{limit:5}),r(re)])])}}});export{be as default};
