import { SignerIface } from 'hypersdk-client/src/client/types'
import { addressHexFromPubKey } from 'hypersdk-client/src/lib/Marshaler.ts'
import { useEffect, useState } from 'react'
import { vmClient } from '../VMClient.ts'
import ConnectWallet from './ConnectWallet'
import Faucet from './Faucet.tsx'
import Wallet from './Wallet'

// Add this type definition at the top of the file
type SignerConnectedEvent = CustomEvent<SignerIface | null>

function App() {
  const [signerConnected, setSignerConnected] = useState<boolean>(false)
  const [myAddr, setMyAddr] = useState<string>('')

  useEffect(() => {
    const handleSignerConnected = (event: SignerConnectedEvent) => {
      setSignerConnected(!!event.detail)
      const signer: SignerIface | null = event.detail
      if (signer) {
        setMyAddr(
          addressHexFromPubKey(signer.getPublicKey()).replace(/^0x/, '')
        )
      } else {
        setMyAddr('')
      }
    }

    vmClient.addEventListener(
      'signerConnected',
      handleSignerConnected as EventListener
    )

    return () =>
      vmClient.removeEventListener(
        'signerConnected',
        handleSignerConnected as EventListener
      )
  }, [])

  if (!signerConnected) {
    return <ConnectWallet />
  }

  return (
    <div className='flex items-center justify-center min-h-screen'>
      <Faucet myAddr={myAddr} minBalance={vmClient.convertToNativeTokens('1')}>
        <Wallet myAddr={myAddr} />
      </Faucet>
    </div>
  )
}

export default App
