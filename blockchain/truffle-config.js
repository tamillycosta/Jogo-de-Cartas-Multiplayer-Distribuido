module.exports = {
  networks: {
    development: {
      host: "127.0.0.1",
      port: 7545,        // porta do ganache-cli
      network_id: "*",
    },
  },

  compilers: {
    solc: {
      version: "0.8.21",
    }
  },

  mocha: {}
};
