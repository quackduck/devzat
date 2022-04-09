import { DtsOptions, RollupOptions } from "dts-cli";
import binary2base64 from "rollup-plugin-binary2base64";

export default {
    rollup(config: RollupOptions, options: DtsOptions) {
        config.plugins.push(binary2base64);
        return config;
    }
}