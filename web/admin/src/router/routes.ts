import type { RouteRecordRaw } from 'vue-router'

export const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/Login.vue'),
    meta: { layout: 'auth', roles: [], title: 'nav.login' },
  },
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    meta: { layout: 'app' },
    children: [
      {
        path: '', name: 'dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { roles: [3], title: 'nav.dashboard', breadcrumb: ['dashboard'] },
      },
      {
        path: 'users', name: 'users',
        component: () => import('@/views/users/List.vue'),
        meta: { roles: [3], title: 'nav.users', breadcrumb: ['users'] },
      },
      {
        path: 'users/:id', name: 'user-detail',
        component: () => import('@/views/users/Detail.vue'),
        meta: { roles: [3], title: 'nav.user_detail', breadcrumb: ['users', ':id'] },
      },
      {
        path: 'groups', name: 'groups',
        component: () => import('@/views/groups/List.vue'),
        meta: { roles: [3], title: 'nav.groups', breadcrumb: ['groups'] },
      },
      {
        path: 'channels', name: 'channels',
        component: () => import('@/views/channels/List.vue'),
        meta: { roles: [3], title: 'nav.channels', breadcrumb: ['channels'] },
      },
      {
        path: 'channels/new', name: 'channel-new',
        component: () => import('@/views/channels/Detail.vue'),
        meta: { roles: [3], title: 'nav.channel_new', breadcrumb: ['channels', 'new'], mode: 'new' },
      },
      {
        path: 'channels/:id', name: 'channel-detail',
        component: () => import('@/views/channels/Detail.vue'),
        meta: { roles: [3], title: 'nav.channel_detail', breadcrumb: ['channels', ':id'] },
      },
      {
        path: 'channels/:id/mappings', name: 'channel-mappings',
        component: () => import('@/views/channels/Mappings.vue'),
        meta: { roles: [3], title: 'nav.channel_mappings', breadcrumb: ['channels', ':id', 'mappings'] },
      },
      {
        path: 'models', name: 'models',
        component: () => import('@/views/models/List.vue'),
        meta: { roles: [3], title: 'nav.models', breadcrumb: ['models'] },
      },
      {
        path: 'pricing', name: 'pricing',
        component: () => import('@/views/pricing/List.vue'),
        meta: { roles: [3], title: 'nav.pricing', breadcrumb: ['pricing'] },
      },
      {
        path: 'tokens', name: 'tokens',
        component: () => import('@/views/tokens/List.vue'),
        meta: { roles: [3], title: 'nav.tokens', breadcrumb: ['tokens'] },
      },
      {
        path: 'logs/requests', name: 'logs-requests',
        component: () => import('@/views/logs/Requests.vue'),
        meta: { roles: [3], title: 'nav.logs_requests', breadcrumb: ['logs', 'requests'] },
      },
      {
        path: 'logs/errors', name: 'logs-errors',
        component: () => import('@/views/logs/Errors.vue'),
        meta: { roles: [3], title: 'nav.logs_errors', breadcrumb: ['logs', 'errors'] },
      },
      {
        path: 'logs/audit', name: 'logs-audit',
        component: () => import('@/views/logs/Audit.vue'),
        meta: { roles: [3], title: 'nav.logs_audit', breadcrumb: ['logs', 'audit'] },
      },
      {
        path: 'stats', name: 'stats',
        component: () => import('@/views/stats/Index.vue'),
        meta: { roles: [3], title: 'nav.stats', breadcrumb: ['stats'] },
      },
      {
        path: 'notices', name: 'notices',
        component: () => import('@/views/notices/List.vue'),
        meta: { roles: [3], title: 'nav.notices', breadcrumb: ['notices'] },
      },
      {
        path: 'notices/new', name: 'notice-new',
        component: () => import('@/views/notices/Detail.vue'),
        meta: { roles: [3], title: 'nav.notice_new', breadcrumb: ['notices', 'new'], mode: 'new' },
      },
      {
        path: 'notices/:id', name: 'notice-detail',
        component: () => import('@/views/notices/Detail.vue'),
        meta: { roles: [3], title: 'nav.notice_detail', breadcrumb: ['notices', ':id'] },
      },
      {
        path: 'payments/recharges', name: 'payments-recharges',
        component: () => import('@/views/payments/Recharges.vue'),
        meta: { roles: [3], title: 'nav.payments_recharges', breadcrumb: ['payments', 'recharges'] },
      },
      {
        path: 'payments/redeem', name: 'payments-redeem',
        component: () => import('@/views/payments/Redeem.vue'),
        meta: { roles: [3], title: 'nav.payments_redeem', breadcrumb: ['payments', 'redeem'] },
      },
      {
        path: 'payments/redeem/batch_new', name: 'payments-redeem-batch',
        component: () => import('@/views/payments/RedeemBatchNew.vue'),
        meta: { roles: [3], title: 'nav.payments_redeem_batch_new', breadcrumb: ['payments', 'redeem', 'batch_new'] },
      },
      {
        path: 'settings', name: 'settings',
        component: () => import('@/views/settings/Index.vue'),
        meta: { roles: [3], title: 'nav.settings', breadcrumb: ['settings'] },
      },
      {
        path: 'profile', name: 'profile',
        component: () => import('@/views/Profile.vue'),
        meta: { roles: [2, 3], title: 'nav.profile', breadcrumb: ['profile'] },
      },
      {
        path: '403', name: 'forbidden',
        component: () => import('@/views/Forbidden.vue'),
        meta: { roles: [], title: 'nav.forbidden' },
      },
      {
        path: '404', name: 'not-found',
        component: () => import('@/views/NotFound.vue'),
        meta: { roles: [], title: 'nav.not_found' },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/404' },
]
