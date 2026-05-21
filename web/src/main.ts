import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { ScadaPlugin } from '@/scada'
import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ScadaPlugin)
app.mount('#app')
