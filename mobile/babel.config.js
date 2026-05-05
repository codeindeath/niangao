// Jest-specific babel config — excludes react-native-reanimated plugin 
// which requires native worklets not available in jest environment.
module.exports = function (api) {
  api.cache(true);
  return {
    presets: [
      ['babel-preset-expo', {
        reanimated: false,  // disable reanimated plugin for jest
      }],
    ],
  };
};
