import { HyperSDKClient } from 'hypersdk-client'
import { HyperSDKHTTPClient } from 'hypersdk-client/dist/HyperSDKHTTPClient'
import { API_HOST, FAUCET_HOST, VM_NAME, VM_RPC_PREFIX } from './const'

export const vmClient = new HyperSDKClient(API_HOST, VM_NAME, VM_RPC_PREFIX)

export async function requestFaucetTransfer(address: string): Promise<void> {
  const response = await fetch(`${FAUCET_HOST}/faucet/${address}`, {
    method: 'POST',
    body: JSON.stringify({})
  })
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }
}

export async function isFaucetReady(): Promise<boolean> {
  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 3000)

    const response = await fetch(`${FAUCET_HOST}/readyz`, {
      method: 'GET',
      signal: controller.signal
    })

    clearTimeout(timeoutId)

    return response.ok
  } catch (error) {
    return false
  }
}

export async function getBalance(
  address: string,
  asset: string
): Promise<bigint> {
  const httpClient = new HyperSDKHTTPClient(API_HOST, VM_NAME, VM_RPC_PREFIX)
  const result = await httpClient.makeVmAPIRequest<{ amount: number }>(
    'balance',
    { address, asset }
  )
  return BigInt(result.amount) // TODO: Handle potential precision loss
}
