import { CapacitorConfig} from "@capacitor/cli";
const config: CapacitorConfig = {
  appId: "io.ionic.demo.pg.vue",
  appName: "Reference App",
  bundledWebRuntime: false,
  npmClient: "npm",
  webDir: "dist",
  plugins: {
    SplashScreen: {
      launchShowDuration: 0
    }
  }
}

export default config
