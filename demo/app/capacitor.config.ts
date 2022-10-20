import { CapacitorConfig} from "@capacitor/cli";
const config: CapacitorConfig = {
  appId: "dev.trustbloc.wallet.demo",
  appName: "TrustBloc DemoApp",
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
