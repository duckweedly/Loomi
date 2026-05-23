import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  base: './',
  plugins: [react()],
  build: {
    rolldownOptions: {
      output: {
        codeSplitting: {
          groups: [
            {
              name: 'react-vendor',
              test: /node_modules[\\/](react|react-dom|scheduler)[\\/]/,
              priority: 40,
            },
            {
              name: 'ui-vendor',
              test: /node_modules[\\/](@lobehub|antd|@ant-design|rc-|@rc-component|classnames|dayjs)[\\/]/,
              priority: 30,
            },
            {
              name: 'icons-vendor',
              test: /node_modules[\\/](lucide-react|@lobehub[\\/]icons|@lobehub[\\/]fluent-emoji)[\\/]/,
              priority: 20,
            },
            {
              name: 'motion-vendor',
              test: /node_modules[\\/](motion|framer-motion)[\\/]/,
              priority: 10,
            },
          ],
        },
      },
    },
  },
})
