// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

import { ArrowPathIcon, ClipboardIcon } from '@heroicons/react/24/outline'
import { addressHexFromPubKey, VMABI } from 'hypersdk-client/dist/Marshaler'
import { stringify } from 'lossless-json'
import { useCallback, useEffect, useReducer, useState } from 'react'
import { getBalance, vmClient } from '../VMClient'

import { hexToBytes } from '@noble/curves/abstract/utils'
import { Block } from 'hypersdk-client/dist/apiTransformers'
import TimeAgo from 'javascript-time-ago'
import timeAgoEn from 'javascript-time-ago/locale/en'
TimeAgo.addDefaultLocale(timeAgoEn)
const ago = new TimeAgo('en-US')

const getDefaultValue = (fieldType: string) => {
  if (fieldType === 'Address')
    return '00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb'
  if (fieldType === '[]uint8') return ''
  if (fieldType === 'string') return ''
  if (fieldType === 'uint64') return '1'
  if (fieldType.startsWith('int') || fieldType.startsWith('uint')) return '0'
  return ''
}

function Action({
  actionName,
  abi,
  fetchBalance
}: {
  actionName: string
  abi: VMABI
  fetchBalance: (waitForChange: boolean) => void
}) {
  const actionType = abi.types.find((t) => t.name === actionName)
  const action = abi.actions.find((a) => a.name === actionName)
  const [actionLogs, setActionLogs] = useState<string[]>([])
  const [actionInputs, setActionInputs] = useState<Record<string, string>>({})
  const [displayInputs, setDisplayInputs] = useState<Record<string, string>>({})

  useEffect(() => {
    if (actionType) {
      setActionInputs((prev) => {
        // Check if actionInputs already have the values to prevent unnecessary updates
        const newActionInputs = { ...prev }
        let needsUpdate = false

        for (const field of actionType.fields) {
          if (!(field.name in newActionInputs)) {
            const defaultValue = getDefaultValue(field.type)
            newActionInputs[field.name] =
              field.type === '[]uint8' ? btoa(defaultValue) : defaultValue
            needsUpdate = true
          }
        }

        // Only update if necessary to avoid infinite loops
        if (needsUpdate) {
          return newActionInputs
        } else {
          return prev
        }
      })

      setDisplayInputs((prev) => {
        // Check if displayInputs already have the values to prevent unnecessary updates
        const newDisplayInputs = { ...prev }
        let needsUpdate = false

        for (const field of actionType.fields) {
          if (!(field.name in newDisplayInputs)) {
            newDisplayInputs[field.name] = getDefaultValue(field.type)
            needsUpdate = true
          }
        }

        // Only update if necessary to avoid infinite loops
        if (needsUpdate) {
          return newDisplayInputs
        } else {
          return prev
        }
      })
    }
    // Remove `displayInputs` and `actionInputs` from the dependency array to avoid infinite loops
  }, [actionName, actionType])

  const executeAction = async (actionName: string, isReadOnly: boolean) => {
    setActionLogs([])
    const now = new Date().toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
    setActionLogs((prev) => [...prev, `${now} - Executing...`])

    // Use displayInputs for logging purposes
    const logInputs = Object.keys(actionInputs).reduce((acc, key) => {
      acc[key] =
        actionInputs[key] === btoa(displayInputs[key])
          ? displayInputs[key]
          : actionInputs[key]
      return acc
    }, {} as Record<string, string>)

    try {
      setActionLogs((prev) => [
        ...prev,
        `Action data for ${actionName}: ${JSON.stringify(logInputs, null, 2)}`
      ])

      // Use the original actionInputs when making the call
      const result = isReadOnly
        ? await vmClient.executeActions([{ actionName, data: actionInputs }])
        : await vmClient.sendTransaction([{ actionName, data: actionInputs }])

      const endTime = new Date().toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
      setActionLogs((prev) => [
        ...prev,
        `${endTime} - Success: ${stringify(result, null, 2)}`
      ])

      if (!isReadOnly) {
        fetchBalance(true)
      }
    } catch (e) {
      console.error(e)
      const errorTime = new Date().toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
      setActionLogs((prev) => [
        ...prev,
        `${errorTime} - Error: ${(e as Error)?.message ?? String(e)}`
      ])
    }
  }

  const handleInputChange = (
    fieldName: string,
    fieldType: string,
    value: string
  ) => {
    setDisplayInputs((prev) => ({
      ...prev,
      [fieldName]: value // Update the displayInputs with the new value
    }))

    setActionInputs((prev) => ({
      ...prev,
      [fieldName]: fieldType === '[]uint8' ? btoa(value) : value // Store transformed value in actionInputs
    }))
  }

  if (!action) {
    return <div>Action not found</div>
  }
  return (
    <div key={action.id} className='mb-6 p-6 bg-white shadow-md rounded-lg'>
      <h3 className='text-2xl font-semibold mb-4 text-gray-800'>
        {action.name}
      </h3>
      <div className='mb-6'>
        <h4 className='font-semibold mb-3 text-gray-700'>Input Fields:</h4>
        {actionType?.fields.map((field) => {
          if (field.type.includes('[]') && field.type !== '[]uint8') {
            return (
              <p key={field.name} className='text-red-500'>
                Warning: Array type not supported for {field.name}
              </p>
            )
          }
          return (
            <div key={field.name} className='mb-4'>
              <label className='block text-sm font-medium text-gray-700 mb-1'>
                {field.name}: {field.type}
              </label>
              <input
                type='text'
                className='mt-1 block w-full rounded-md border border-gray-200 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 p-2'
                value={displayInputs[field.name] ?? ''}
                onChange={(e) =>
                  handleInputChange(field.name, field.type, e.target.value)
                }
              />
            </div>
          )
        })}
      </div>
      <div className='flex space-x-4 mb-4'>
        <button
          onClick={() => executeAction(action.name, true)}
          className='px-4 py-2 bg-gray-200 text-gray-800 font-bold rounded-md hover:bg-gray-300 transition duration-300'
        >
          Execute Read Only
        </button>
        <button
          onClick={() => executeAction(action.name, false)}
          className='px-4 py-2 bg-black text-white font-bold rounded-md hover:bg-gray-800 transition duration-300'
        >
          Execute in Transaction
        </button>
      </div>
      <div className='bg-gray-100 p-4 rounded-md'>
        <pre className='whitespace-pre-wrap text-sm'>
          {actionLogs.join('\n') || 'No logs yet.'}
        </pre>
      </div>
    </div>
  )
}

export default function Wallet({ myAddr }: { myAddr: string }) {
  const [balance, setBalance] = useState<bigint | null>(null)
  const [balanceLoading, setBalanceLoading] = useState(false)
  const [balanceError, setBalanceError] = useState<string | null>(null)
  const [abi, setAbi] = useState<VMABI | null>(null)
  const [abiLoading, setAbiLoading] = useState(false)
  const [abiError, setAbiError] = useState<string | null>(null)

  // Balance fetching
  const fetchBalance = useCallback(async () => {
    setBalanceLoading(true)
    try {
      setBalance(await getBalance(myAddr, 'NAI'))
      setBalanceError(null)
    } catch (e) {
      console.error('Failed to fetch balance:', e)
      setBalanceError(e instanceof Error && e.message ? e.message : String(e))
    } finally {
      setBalanceLoading(false)
    }
  }, [myAddr])

  useEffect(() => {
    fetchBalance()
  }, [fetchBalance])

  // ABI fetching
  useEffect(() => {
    setAbiLoading(true)
    vmClient
      .getAbi()
      .then((newAbi) => {
        setAbi(newAbi)
        setAbiError(null)
      })
      .catch((e) => {
        console.error('Failed to fetch ABI:', e)
        setAbiError(e instanceof Error && e.message ? e.message : String(e))
      })
      .finally(() => setAbiLoading(false))
  }, [])

  const copyToClipboard = () => {
    navigator.clipboard
      .writeText(myAddr)
      .then(() => {
        // You can add a notification here if you want
        console.log('Address copied to clipboard')
      })
      .catch((err) => {
        console.error('Failed to copy address: ', err)
      })
  }

  return (
    <div className='w-full bg-gray-100 p-8'>
      <div className='lg:flex lg:space-x-8'>
        <div className='lg:w-2/3'>
          <h2 className='text-2xl font-semibold mb-4 text-gray-800'>
            My Wallet
          </h2>
          <div className='mb-8 bg-white p-6 rounded-lg shadow-md'>
            <div className='flex items-center justify-between'>
              <div className='font-mono break-all py-4 rounded-md'>
                {myAddr}
              </div>
              <button
                onClick={copyToClipboard}
                className='p-2 rounded-full hover:bg-gray-200 transition duration-300'
              >
                <ClipboardIcon className='h-6 w-6 text-gray-600' />
              </button>
            </div>
            <div>
              {balanceLoading ? (
                <div className='text-gray-600'>Loading balance...</div>
              ) : balanceError ? (
                <div className='text-red-500'>Error: {balanceError}</div>
              ) : balance !== null ? (
                <div className='flex items-center justify-between'>
                  <div className='text-4xl font-bold text-gray-800'>
                    {parseFloat(vmClient.formatNativeTokens(balance)).toFixed(
                      6
                    )}{' '}
                    NAI
                  </div>
                  <button
                    onClick={() => fetchBalance()}
                    className='p-2 rounded-full hover:bg-gray-200 transition duration-300'
                  >
                    <ArrowPathIcon className='h-6 w-6 text-gray-600' />
                  </button>
                </div>
              ) : null}
            </div>
          </div>
          <div>
            {abiLoading ? (
              <div className='text-gray-600'>Loading ABI...</div>
            ) : abiError ? (
              <div className='text-red-500'>Error loading ABI: {abiError}</div>
            ) : abi ? (
              abi.actions.map((action) => (
                <Action
                  key={action.id}
                  actionName={action.name}
                  abi={abi}
                  fetchBalance={fetchBalance}
                />
              ))
            ) : null}
          </div>
        </div>
        <LatestBlocks />
      </div>
    </div>
  )
}

export function LatestBlocks() {
  const [blocks, setBlocks] = useState([] as Block[])
  const [, forceUpdate] = useReducer((x: number) => x + 1, 0)

  useEffect(() => {
    const unsubscribe = vmClient.listenToBlocks((block) => {
      console.log('New block', block)
      setBlocks((prevBlocks) => [block, ...prevBlocks].slice(0, 5))
    })

    const intervalId = setInterval(() => {
      forceUpdate()
    }, 10000)

    return () => {
      // eslint-disable-next-line no-extra-semi
      ;(async () => {
        try {
          // eslint-disable-next-line no-extra-semi
          ;(await unsubscribe)()
          clearInterval(intervalId)
        } catch (error) {
          console.error('Error unsubscribing:', error)
        }
      })()
    }
  }, [])

  return (
    <div className='lg:w-1/3 mt-8 lg:mt-0'>
      <h2 className='text-2xl font-semibold mb-4 text-gray-800'>
        Latest Blocks
      </h2>
      <div>
        {blocks.length === 0 ? (
          <p className='text-gray-600'>
            Waiting for new blocks (empty blocks are skipped)...
          </p>
        ) : (
          blocks.map((block) => (
            <RenderBlock key={block.block.height} block={block} />
          ))
        )}
      </div>
    </div>
  )
}

function RenderBlock({ block }: { block: Block }) {
  const [showFullJson, setShowFullJson] = useState(false)

  return (
    <div
      key={block.block.height}
      className='mb-6 last:mb-0 bg-white p-6 rounded-lg shadow-md'
    >
      <h3 className='text-2xl font-bold text-gray-800'>
        Block #{block.block.height}
      </h3>
      <p className='text-sm text-gray-600 mb-3'>
        {ago.format(block.block.timestamp, 'round')}
      </p>
      <div className='mb-4'>
        <p className='text-xs text-gray-500 truncate'>
          Parent: <span className='font-mono'>{block.block.parent}</span>
        </p>
        <p className='text-xs text-gray-500 truncate'>
          State Root: <span className='font-mono'>{block.block.stateRoot}</span>
        </p>
      </div>
      <p className='text-sm font-semibold mb-3 text-gray-700'>
        {block.block.txs.length} Transaction
        {block.block.txs.length === 1 ? '' : 's'}
      </p>
      {block.block.txs.map((tx, index) => (
        <div key={index}>
          <p className='font-semibold text-gray-800'>
            {block.results[index].success ? '✅ Success' : '❌ Failed'}
          </p>
          <p className='text-xs mt-2 text-gray-600'>
            Sender:{' '}
            <span className='font-mono'>
              {addressHexFromPubKey(hexToBytes(tx.auth.signer))}
            </span>
          </p>

          <div
            key={index}
            className='mt-4 p-4 bg-gray-100 rounded-md shadow-sm'
          >
            <div className='mt-3 overflow-x-auto'>
              <pre className='text-xs text-gray-700'>
                {stringify(
                  {
                    actions: tx.actions,
                    outputs: block.results[index].outputs
                  },
                  null,
                  2
                )}
              </pre>
            </div>
          </div>
        </div>
      ))}
      <div className='mt-4'>
        <button
          onClick={() => setShowFullJson(!showFullJson)}
          className='text-blue-600 hover:text-blue-800 transition duration-300'
        >
          {showFullJson ? 'Hide' : 'Show'} Full JSON
        </button>
        {showFullJson && (
          <div className='mt-2 bg-gray-100 p-4 rounded-md shadow-sm overflow-x-auto'>
            <pre className='text-xs text-gray-700'>
              {stringify(block, null, 2)}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}
