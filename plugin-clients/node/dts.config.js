const { readFile } = require("fs/promises");
const { exec } = require("child_process");

module.exports = {
    rollup(config, options) {
        config.plugins.push({
            name: "binary",
            async transform(code, id) {
                if(id.endsWith(".pb")) {
                    // it tried to read code as a string, but it's actually binary data
                    // I can't figure out how to turn it back into a buffer properly, so just reread the file
                    const bin = await readFile(id);
                    return `export default ${JSON.stringify(bin.toString("base64"))}`;
                }
            }
        });
        config.plugins.push({
            name: "copy-dts",
            async buildEnd() {
                await exec("yarn copy-dts");
            }
        })
        return config;
    }
}