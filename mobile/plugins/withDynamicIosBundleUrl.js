const { withDangerousMod } = require('@expo/config-plugins');
const fs = require('fs');
const path = require('path');

const DYNAMIC_BUNDLE_URL =
  'return RCTBundleURLProvider.sharedSettings().jsBundleURL(forBundleRoot: "index")';

module.exports = function withDynamicIosBundleUrl(config) {
  return withDangerousMod(config, [
    'ios',
    async (modConfig) => {
      const appDelegatePath = path.join(
        modConfig.modRequest.platformProjectRoot,
        modConfig.modRequest.projectName,
        'AppDelegate.swift'
      );

      if (!fs.existsSync(appDelegatePath)) {
        return modConfig;
      }

      const source = fs.readFileSync(appDelegatePath, 'utf8');
      if (source.includes(DYNAMIC_BUNDLE_URL)) {
        return modConfig;
      }

      const updated = source.replace(
        /return URL\(string: "http:\/\/[^"]+:8081\/index\.bundle\?platform=ios&dev=true&minify=false"\)!/,
        DYNAMIC_BUNDLE_URL
      );

      if (updated !== source) {
        fs.writeFileSync(appDelegatePath, updated);
      }

      return modConfig;
    },
  ]);
};
