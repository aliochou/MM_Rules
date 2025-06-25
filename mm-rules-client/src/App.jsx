import React, { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'

const BACKEND_URL = 'http://localhost:8080'
const GAME_ID = 'demo-game'
const PLAYER_ID = 'demo-player'

function App() {
  const [status, setStatus] = useState('')
  const [requestId, setRequestId] = useState(null)
  const [session, setSession] = useState(null)
  const [loading, setLoading] = useState(false)

  const createMatch = async () => {
    const playerId = 'demo-player-' + Math.floor(Math.random() * 100000);
    setStatus('Creating match request...')
    setSession(null)
    setLoading(true)
    setRequestId(null)
    try {
      // Step 1: Create match request
      const res = await fetch(`${BACKEND_URL}/api/v1/match-request`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ player_id: playerId, game_id: GAME_ID, metadata: {} }),
      })
      const data = await res.json()
      if (!data.request_id) throw new Error('No request_id returned')
      setRequestId(data.request_id)
      setStatus('Waiting for match...')
      // Step 2: Trigger matchmaking processing
      await fetch(`${BACKEND_URL}/api/v1/process-matchmaking/${GAME_ID}`, { method: 'POST' });
      // Step 3: Poll for match status
      pollStatus(data.request_id, 0)
    } catch (err) {
      setStatus('Error: ' + err.message)
      setLoading(false)
    }
  }

  const pollStatus = async (reqId, attempt) => {
    if (attempt > 30) {
      setStatus('Timed out waiting for match.')
      setLoading(false)
      return
    }
    try {
      const res = await fetch(`${BACKEND_URL}/api/v1/match-status/${reqId}`)
      const data = await res.json()
      if (data.status === 'matched' || data.status === 'allocated') {
        setStatus('Match found!')
        if (data.session) {
          setSession(data.session)
          setStatus('Match is running!')
        }
        setLoading(false)
        return
      }
      setTimeout(() => pollStatus(reqId, attempt + 1), 1000)
    } catch (err) {
      setStatus('Error polling status: ' + err.message)
      setLoading(false)
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', marginTop: 80 }}>
      <h1>MM-Rules Demo Client</h1>
      <button onClick={createMatch} disabled={loading} style={{ fontSize: 24, padding: '12px 32px' }}>
        {loading ? 'Processing...' : 'Create Match'}
      </button>
      <div style={{ marginTop: 32, fontSize: 20 }}>{status}</div>
      {session && (
        <div style={{ marginTop: 16, fontSize: 18 }}>
          <div>Session IP: {session.ip}</div>
          <div>Session Port: {session.port}</div>
          <div>Session ID: {session.id}</div>
        </div>
      )}
    </div>
  )
}

export default App
