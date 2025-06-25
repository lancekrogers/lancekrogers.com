module.exports = {
  rootDir: '../../',
  testEnvironment: 'node', // e2e tests run in node environment with puppeteer
  transform: {
    '^.+\\.js$': ['babel-jest', { 
      presets: [
        ['@babel/preset-env', { targets: { node: 'current' } }]
      ] 
    }]
  },
  testMatch: [
    '<rootDir>/tests/js/e2e.test.js'
  ],
  verbose: true,
  testTimeout: 30000, // Longer timeout for e2e tests
  testPathIgnorePatterns: ['/node_modules/', '/test_node_modules/']
};