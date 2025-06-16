module.exports = {
  testEnvironment: 'jsdom',
  transform: {
    '^.+\\.js$': ['babel-jest', { 
      presets: [
        ['@babel/preset-env', { targets: { node: 'current' } }]
      ] 
    }]
  },
  moduleNameMapper: {
    '\\.(css|less|scss|sass)$': '<rootDir>/test-setup.js',
  },
  setupFilesAfterEnv: ['<rootDir>/test-setup.js'],
  testMatch: [
    '**/*.test.js',
    '!**/e2e.test.js'
  ],
  collectCoverageFrom: [
    'static/**/*.js',
    '!static/**/*.test.js',
    '!node_modules/**',
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70
    }
  },
  verbose: true,
  testPathIgnorePatterns: ['/node_modules/', '/test_node_modules/']
};