// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import App from './App'

import Explorer from './components/Explorer'
import Faucet from './components/Faucet'
import Feed from './components/Feed'
import Mint from './components/Mint'
import Transfer from './components/Transfer'


const routes = [
  {
    path: '/',
    element: <App />,
    children: [
      { index: true, element: <Explorer /> },
      {
        path: 'explorer',
        element: <Explorer />
      },
      {
        path: 'faucet',
        element: <Faucet />
      },
      {
        path: 'mint',
        element: <Mint />
      },
      {
        path: 'transfer',
        element: <Transfer />
      },
      {
        path: 'feed',
        element: <Feed />
      }
    ]
  }
]

export default routes
