import React, { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'

type Idea = { id: string; title: string; description: string; userEmail: string; createdAt: string }
const API_URL = '/graphql'

async function gql(query: string, variables: Record<string, unknown> = {}, token?: string) {
  const res = await fetch(API_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: JSON.stringify({ query, variables })
  })
  const data = await res.json()
  if (data.errors?.length) throw new Error(data.errors[0].message)
  return data.data
}

function App() {
  const [token, setToken] = useState(localStorage.getItem('token') || '')
  const [email, setEmail] = useState(localStorage.getItem('email') || '')
  const [ideas, setIdeas] = useState<Idea[]>([])
  const [authEmail, setAuthEmail] = useState('')
  const [password, setPassword] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')

  const load = async () => {
    const d = await gql('query { ideas { id title description userEmail createdAt } }')
    setIdeas(d.ideas)
  }
  useEffect(() => { load().catch((e) => setError(String(e.message || e))) }, [])

  const auth = async (mode: 'register' | 'login') => {
    setError('')
    const m = mode === 'register'
      ? 'mutation($email:String!,$password:String!){ register(email:$email,password:$password){ token email } }'
      : 'mutation($email:String!,$password:String!){ login(email:$email,password:$password){ token email } }'
    const d = await gql(m, { email: authEmail, password })
    const p = mode === 'register' ? d.register : d.login
    setToken(p.token); setEmail(p.email)
    localStorage.setItem('token', p.token); localStorage.setItem('email', p.email)
  }

  const createIdea = async () => {
    await gql('mutation($title:String!,$description:String!){ createIdea(title:$title,description:$description){ id } }', { title, description }, token)
    setTitle(''); setDescription(''); await load()
  }

  return <main style={{ maxWidth: 900, margin: '2rem auto', fontFamily: 'system-ui', padding: 12 }}>
    <h1>Project Ideas Portal</h1>
    {error && <p style={{ color: 'crimson' }}>{error}</p>}
    {!token ? <section>
      <input placeholder='Email' value={authEmail} onChange={e => setAuthEmail(e.target.value)} />{' '}
      <input placeholder='Password' type='password' value={password} onChange={e => setPassword(e.target.value)} />{' '}
      <button onClick={() => auth('register')}>Register</button>{' '}
      <button onClick={() => auth('login')}>Login</button>
    </section> : <section>
      <p>Logged in as <b>{email}</b> <button onClick={() => { setToken(''); setEmail(''); localStorage.clear() }}>Logout</button></p>
      <input placeholder='Idea title' value={title} onChange={e => setTitle(e.target.value)} style={{ width: '100%', marginBottom: 8 }} />
      <textarea placeholder='Idea description' value={description} onChange={e => setDescription(e.target.value)} style={{ width: '100%', minHeight: 100 }} />
      <br /><button onClick={createIdea}>Submit idea</button>
    </section>}
    <h2>Ideas</h2>
    <ul>{ideas.map(i => <li key={i.id}><b>{i.title}</b> by {i.userEmail}<br />{i.description}</li>)}</ul>
  </main>
}

createRoot(document.getElementById('root')!).render(<App />)
