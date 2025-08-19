/**
 * Blockchain-Based Loyalty Program System
 * 
 * Implements a decentralized loyalty program using blockchain technology
 * with smart contracts, NFT rewards, and cryptocurrency integration
 */

export interface LoyaltyToken {
  id: string
  user_id: string
  token_type: 'TRAVEL_POINTS' | 'EXPERIENCE_NFT' | 'MILESTONE_BADGE' | 'CASHBACK_TOKEN'
  amount: number
  metadata: {
    name: string
    description: string
    image_url?: string
    attributes: Record<string, any>
    rarity: 'common' | 'rare' | 'epic' | 'legendary'
  }
  blockchain_address: string
  transaction_hash: string
  created_at: string
  expires_at?: string
}

export interface LoyaltyTransaction {
  id: string
  user_id: string
  type: 'earn' | 'redeem' | 'transfer' | 'stake'
  token_type: string
  amount: number
  source: string
  destination?: string
  transaction_hash: string
  block_number: number
  gas_fee: number
  status: 'pending' | 'confirmed' | 'failed'
  created_at: string
}

export interface SmartContract {
  address: string
  abi: any[]
  network: 'ethereum' | 'polygon' | 'binance' | 'avalanche'
  gas_limit: number
  gas_price: number
}

export interface NFTReward {
  token_id: string
  contract_address: string
  name: string
  description: string
  image_url: string
  attributes: Array<{
    trait_type: string
    value: string | number
    rarity_score: number
  }>
  rarity_rank: number
  unlock_condition: string
  utility: string[]
  transferable: boolean
}

export interface StakingPool {
  id: string
  name: string
  token_type: string
  apy: number
  min_stake: number
  lock_period_days: number
  total_staked: number
  rewards_distributed: number
  is_active: boolean
}

export interface WalletConnection {
  address: string
  network: string
  provider: 'metamask' | 'walletconnect' | 'coinbase' | 'phantom'
  balance: Record<string, number>
  is_connected: boolean
}

class BlockchainLoyaltyService {
  private static readonly CONTRACTS: Record<string, SmartContract> = {
    TRAVEL_POINTS: {
      address: '0x1234567890123456789012345678901234567890',
      abi: [], // Smart contract ABI would go here
      network: 'polygon',
      gas_limit: 100000,
      gas_price: 20
    },
    NFT_REWARDS: {
      address: '0x0987654321098765432109876543210987654321',
      abi: [], // NFT contract ABI would go here
      network: 'polygon',
      gas_limit: 150000,
      gas_price: 25
    }
  }

  private static wallet: WalletConnection | null = null
  private static web3Provider: any = null

  // Wallet Connection
  static async connectWallet(provider: 'metamask' | 'walletconnect' | 'coinbase' = 'metamask'): Promise<WalletConnection> {
    try {
      if (typeof window === 'undefined') {
        throw new Error('Wallet connection only available in browser')
      }

      let ethereum: any
      
      switch (provider) {
        case 'metamask':
          ethereum = (window as any).ethereum
          if (!ethereum) {
            throw new Error('MetaMask not installed')
          }
          break
        case 'walletconnect':
          // WalletConnect integration would go here
          throw new Error('WalletConnect not implemented yet')
        case 'coinbase':
          // Coinbase Wallet integration would go here
          throw new Error('Coinbase Wallet not implemented yet')
      }

      // Request account access
      const accounts = await ethereum.request({ method: 'eth_requestAccounts' })
      const chainId = await ethereum.request({ method: 'eth_chainId' })
      
      // Get network name
      const networkMap: Record<string, string> = {
        '0x1': 'ethereum',
        '0x89': 'polygon',
        '0x38': 'binance',
        '0xa86a': 'avalanche'
      }
      
      const network = networkMap[chainId] || 'unknown'
      
      // Get balances
      const balance = await this.getWalletBalances(accounts[0], network)
      
      this.wallet = {
        address: accounts[0],
        network,
        provider,
        balance,
        is_connected: true
      }

      // Store connection in localStorage
      localStorage.setItem('wallet_connection', JSON.stringify(this.wallet))
      
      return this.wallet
      
    } catch (error) {
      console.error('Wallet connection failed:', error)
      throw error
    }
  }

  static async disconnectWallet(): Promise<void> {
    this.wallet = null
    this.web3Provider = null
    localStorage.removeItem('wallet_connection')
  }

  static getConnectedWallet(): WalletConnection | null {
    if (this.wallet) return this.wallet
    
    // Try to restore from localStorage
    try {
      const stored = localStorage.getItem('wallet_connection')
      if (stored) {
        this.wallet = JSON.parse(stored)
        return this.wallet
      }
    } catch (error) {
      console.error('Error restoring wallet connection:', error)
    }
    
    return null
  }

  // Token Management
  static async earnTokens(
    userId: string, 
    tokenType: LoyaltyToken['token_type'], 
    amount: number, 
    source: string,
    metadata?: Partial<LoyaltyToken['metadata']>
  ): Promise<LoyaltyToken> {
    try {
      if (!this.wallet) {
        throw new Error('Wallet not connected')
      }

      // Simulate smart contract interaction
      const transactionHash = await this.simulateTransaction('mint', {
        to: this.wallet.address,
        tokenType,
        amount
      })

      const token: LoyaltyToken = {
        id: `token_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        user_id: userId,
        token_type: tokenType,
        amount,
        metadata: {
          name: this.getTokenName(tokenType),
          description: this.getTokenDescription(tokenType, source),
          rarity: this.calculateRarity(amount, tokenType),
          attributes: {
            source,
            earned_at: new Date().toISOString(),
            ...metadata
          }
        },
        blockchain_address: this.wallet.address,
        transaction_hash: transactionHash,
        created_at: new Date().toISOString()
      }

      // Store token locally (in real app, this would be on blockchain)
      this.storeToken(token)
      
      return token
      
    } catch (error) {
      console.error('Error earning tokens:', error)
      throw error
    }
  }

  static async redeemTokens(
    userId: string,
    tokenType: LoyaltyToken['token_type'],
    amount: number,
    rewardId: string
  ): Promise<LoyaltyTransaction> {
    try {
      if (!this.wallet) {
        throw new Error('Wallet not connected')
      }

      // Check balance
      const balance = await this.getTokenBalance(userId, tokenType)
      if (balance < amount) {
        throw new Error('Insufficient token balance')
      }

      // Simulate smart contract interaction
      const transactionHash = await this.simulateTransaction('burn', {
        from: this.wallet.address,
        tokenType,
        amount
      })

      const transaction: LoyaltyTransaction = {
        id: `tx_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        user_id: userId,
        type: 'redeem',
        token_type: tokenType,
        amount,
        source: this.wallet.address,
        destination: rewardId,
        transaction_hash: transactionHash,
        block_number: Math.floor(Math.random() * 1000000),
        gas_fee: 0.001,
        status: 'confirmed',
        created_at: new Date().toISOString()
      }

      // Store transaction
      this.storeTransaction(transaction)
      
      return transaction
      
    } catch (error) {
      console.error('Error redeeming tokens:', error)
      throw error
    }
  }

  // NFT Rewards
  static async mintNFTReward(
    userId: string,
    rewardType: string,
    unlockCondition: string
  ): Promise<NFTReward> {
    try {
      if (!this.wallet) {
        throw new Error('Wallet not connected')
      }

      const tokenId = `nft_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      
      // Simulate NFT minting
      const transactionHash = await this.simulateTransaction('mintNFT', {
        to: this.wallet.address,
        tokenId,
        metadata: {
          name: `${rewardType} Achievement`,
          description: `Unlocked by: ${unlockCondition}`
        }
      })

      const nftReward: NFTReward = {
        token_id: tokenId,
        contract_address: this.CONTRACTS.NFT_REWARDS.address,
        name: `${rewardType} Achievement NFT`,
        description: `Exclusive NFT reward for ${unlockCondition}`,
        image_url: `/nft-rewards/${rewardType.toLowerCase()}.png`,
        attributes: [
          {
            trait_type: 'Achievement Type',
            value: rewardType,
            rarity_score: this.calculateNFTRarity(rewardType)
          },
          {
            trait_type: 'Unlock Date',
            value: new Date().toISOString().split('T')[0],
            rarity_score: 1
          }
        ],
        rarity_rank: Math.floor(Math.random() * 1000) + 1,
        unlock_condition: unlockCondition,
        utility: this.getNFTUtility(rewardType),
        transferable: true
      }

      // Store NFT locally
      this.storeNFT(userId, nftReward)
      
      return nftReward
      
    } catch (error) {
      console.error('Error minting NFT reward:', error)
      throw error
    }
  }

  // Staking
  static async stakeTokens(
    userId: string,
    poolId: string,
    tokenType: string,
    amount: number
  ): Promise<LoyaltyTransaction> {
    try {
      if (!this.wallet) {
        throw new Error('Wallet not connected')
      }

      const pool = await this.getStakingPool(poolId)
      if (!pool || !pool.is_active) {
        throw new Error('Staking pool not available')
      }

      if (amount < pool.min_stake) {
        throw new Error(`Minimum stake amount is ${pool.min_stake}`)
      }

      // Simulate staking transaction
      const transactionHash = await this.simulateTransaction('stake', {
        from: this.wallet.address,
        poolId,
        tokenType,
        amount
      })

      const transaction: LoyaltyTransaction = {
        id: `stake_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        user_id: userId,
        type: 'stake',
        token_type: tokenType,
        amount,
        source: this.wallet.address,
        destination: poolId,
        transaction_hash: transactionHash,
        block_number: Math.floor(Math.random() * 1000000),
        gas_fee: 0.002,
        status: 'confirmed',
        created_at: new Date().toISOString()
      }

      this.storeTransaction(transaction)
      return transaction
      
    } catch (error) {
      console.error('Error staking tokens:', error)
      throw error
    }
  }

  // Helper Methods
  private static async getWalletBalances(address: string, network: string): Promise<Record<string, number>> {
    // Simulate getting balances from blockchain
    return {
      ETH: Math.random() * 10,
      MATIC: Math.random() * 1000,
      TRAVEL_POINTS: Math.floor(Math.random() * 10000),
      EXPERIENCE_NFT: Math.floor(Math.random() * 50)
    }
  }

  private static async simulateTransaction(method: string, params: any): Promise<string> {
    // Simulate blockchain transaction
    await new Promise(resolve => setTimeout(resolve, 2000))
    return `0x${Math.random().toString(16).substr(2, 64)}`
  }

  private static getTokenName(tokenType: LoyaltyToken['token_type']): string {
    const names = {
      TRAVEL_POINTS: 'Travel Points',
      EXPERIENCE_NFT: 'Experience NFT',
      MILESTONE_BADGE: 'Milestone Badge',
      CASHBACK_TOKEN: 'Cashback Token'
    }
    return names[tokenType]
  }

  private static getTokenDescription(tokenType: LoyaltyToken['token_type'], source: string): string {
    return `${this.getTokenName(tokenType)} earned from ${source}`
  }

  private static calculateRarity(amount: number, tokenType: LoyaltyToken['token_type']): 'common' | 'rare' | 'epic' | 'legendary' {
    if (amount >= 10000) return 'legendary'
    if (amount >= 5000) return 'epic'
    if (amount >= 1000) return 'rare'
    return 'common'
  }

  private static calculateNFTRarity(rewardType: string): number {
    const rarityMap: Record<string, number> = {
      'First Booking': 10,
      'Frequent Traveler': 25,
      'Explorer': 50,
      'Adventurer': 75,
      'Globe Trotter': 100
    }
    return rarityMap[rewardType] || 1
  }

  private static getNFTUtility(rewardType: string): string[] {
    const utilityMap: Record<string, string[]> = {
      'First Booking': ['5% discount on next booking'],
      'Frequent Traveler': ['Priority customer support', '10% discount'],
      'Explorer': ['Access to exclusive destinations', '15% discount'],
      'Adventurer': ['VIP lounge access', '20% discount'],
      'Globe Trotter': ['Personal travel concierge', '25% discount', 'Free upgrades']
    }
    return utilityMap[rewardType] || []
  }

  // Storage Methods (in real app, these would interact with blockchain)
  private static storeToken(token: LoyaltyToken): void {
    const stored = JSON.parse(localStorage.getItem('loyalty_tokens') || '[]')
    stored.push(token)
    localStorage.setItem('loyalty_tokens', JSON.stringify(stored))
  }

  private static storeTransaction(transaction: LoyaltyTransaction): void {
    const stored = JSON.parse(localStorage.getItem('loyalty_transactions') || '[]')
    stored.push(transaction)
    localStorage.setItem('loyalty_transactions', JSON.stringify(stored))
  }

  private static storeNFT(userId: string, nft: NFTReward): void {
    const stored = JSON.parse(localStorage.getItem(`nft_rewards_${userId}`) || '[]')
    stored.push(nft)
    localStorage.setItem(`nft_rewards_${userId}`, JSON.stringify(stored))
  }

  // Query Methods
  static async getTokenBalance(userId: string, tokenType: LoyaltyToken['token_type']): Promise<number> {
    const tokens = JSON.parse(localStorage.getItem('loyalty_tokens') || '[]')
    return tokens
      .filter((t: LoyaltyToken) => t.user_id === userId && t.token_type === tokenType)
      .reduce((sum: number, t: LoyaltyToken) => sum + t.amount, 0)
  }

  static async getUserNFTs(userId: string): Promise<NFTReward[]> {
    return JSON.parse(localStorage.getItem(`nft_rewards_${userId}`) || '[]')
  }

  static async getStakingPool(poolId: string): Promise<StakingPool | null> {
    // Mock staking pools
    const pools: StakingPool[] = [
      {
        id: 'travel_points_pool',
        name: 'Travel Points Staking',
        token_type: 'TRAVEL_POINTS',
        apy: 12.5,
        min_stake: 1000,
        lock_period_days: 30,
        total_staked: 500000,
        rewards_distributed: 62500,
        is_active: true
      }
    ]
    
    return pools.find(p => p.id === poolId) || null
  }

  static async getUserTransactions(userId: string): Promise<LoyaltyTransaction[]> {
    const transactions = JSON.parse(localStorage.getItem('loyalty_transactions') || '[]')
    return transactions.filter((t: LoyaltyTransaction) => t.user_id === userId)
  }

  static async getUserTokens(userId: string): Promise<LoyaltyToken[]> {
    const tokens = JSON.parse(localStorage.getItem('loyalty_tokens') || '[]')
    return tokens.filter((t: LoyaltyToken) => t.user_id === userId)
  }
}

export { BlockchainLoyaltyService }
