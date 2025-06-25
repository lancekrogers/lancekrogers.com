// Blockchain Animation Components (Disabled)
// These can be re-enabled on other pages by including this file and calling the functions

function createMempoolTransactions(container) {
  const mempoolContainer = document.createElement('div');
  mempoolContainer.className = 'mempool-container';
  
  // Create individual transaction elements that will converge into a block
  const transactions = [
    { hash: 'a1b2c3...', value: '0.15 ETH', delay: 0 },
    { hash: 'f4e5d6...', value: '2.3 ETH', delay: 0.5 },
    { hash: '9g8h7i...', value: '0.8 ETH', delay: 1 },
    { hash: 'k2l3m4...', value: '1.1 ETH', delay: 1.5 },
    { hash: 'n5o6p7...', value: '0.45 ETH', delay: 2 }
  ];
  
  transactions.forEach((tx, index) => {
    const txElement = document.createElement('div');
    txElement.className = 'mempool-transaction';
    txElement.innerHTML = `
      <span class="tx-hash">${tx.hash}</span>
      <span class="tx-value">${tx.value}</span>
    `;
    txElement.style.animationDelay = `${tx.delay}s`;
    txElement.dataset.txIndex = index;
    mempoolContainer.appendChild(txElement);
  });
  
  // Create the block that forms after transactions converge
  const blockElement = document.createElement('div');
  blockElement.className = 'blockchain-block';
  blockElement.innerHTML = `
    <div class="block-header">Block #18,742,156</div>
    <div class="block-hash">0x4a2b8c...</div>
    <div class="block-txs">${transactions.length} txs</div>
  `;
  mempoolContainer.appendChild(blockElement);
  
  container.appendChild(mempoolContainer);
}

function createEthereumContract(container) {
  const contractElement = document.createElement('div');
  contractElement.className = 'ethereum-contract';
  contractElement.innerHTML = `
    <div class="contract-header">Contract Deployed</div>
    <div class="contract-address">0x742d35Cc6C4...</div>
    <div class="contract-bytecode">
      <span>6080604052348015600f57600080fd5b50...</span>
    </div>
  `;
  
  container.appendChild(contractElement);
}

// Export functions for potential use on other pages
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    createMempoolTransactions,
    createEthereumContract
  };
}