import DefaultTheme from 'vitepress/theme'
import { withBase } from 'vitepress'
import './styles.css'

export default {
  ...DefaultTheme,
  enhanceApp({ app }: { app: any }) {
    app.config.globalProperties.$withBase = withBase
  },
}
