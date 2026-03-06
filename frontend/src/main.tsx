import React, { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import './styles.css'

type Idea = { id: string; title: string; description: string; userEmail: string; createdAt: string }
const API_URL = '/graphql'

async function gql(query: string, variables: Record<string, unknown> = {}, token?: string) {
  const response = await fetch(API_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
    body: JSON.stringify({ query, variables })
  })
  const payload = await response.json()
  if (payload.errors?.length) throw new Error(payload.errors[0].message)
  return payload.data
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

  const loadIdeas = async () => {
    const data = await gql(`query { ideas { id title description userEmail createdAt } }`)
    setIdeas(data.ideas)
  }

  useEffect(() => {
    loadIdeas().catch((e) => setError(String(e.message || e)))
  }, [])

  const handleAuth = async (mode: 'register' | 'login') => {
    setError('')
    const mutation =
      mode === 'register'
        ? `mutation($email:String!,$password:String!){ register(email:$email,password:$password){ token email } }`
        : `mutation($email:String!,$password:String!){ login(email:$email,password:$password){ token email } }`
    const data = await gql(mutation, { email: authEmail, password })
    const auth = mode === 'register' ? data.register : data.login
    setToken(auth.token)
    setEmail(auth.email)
    localStorage.setItem('token', auth.token)
    localStorage.setItem('email', auth.email)
  }

  const submitIdea = async () => {
    setError('')
    await gql(
      `mutation($title:String!,$description:String!){ createIdea(title:$title,description:$description){ id } }`,
      { title, description },
      token,
    )
    setTitle('')
    setDescription('')
    await loadIdeas()
  }

  const logout = () => {
    setToken('')
    setEmail('')
    localStorage.removeItem('token')
    localStorage.removeItem('email')
  }

  return (
    <main className="mx-auto max-w-3xl p-4 sm:p-8">
      <h1 className="text-3xl font-bold text-slate-900">Project Ideas Portal</h1>
      <p className="mt-2 text-slate-700">Register or login, then submit project ideas.</p>

      {error && <p className="mt-4 rounded bg-red-100 p-3 text-red-700">{error}</p>}

      <section className="mt-6 rounded-xl bg-white p-4 shadow">
        {!token ? (
          <>
            <h2 className="text-lg font-semibold">Account</h2>
            <div className="mt-3 grid gap-2 sm:grid-cols-2">
              <input className="rounded border p-2" placeholder="Email" value={authEmail} onChange={(e) => setAuthEmail(e.target.value)} />
              <input className="rounded border p-2" placeholder="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </div>
            <div className="mt-3 flex gap-2">
              <button className="rounded bg-slate-900 px-3 py-2 text-white" onClick={() => handleAuth('register')}>Register</button>
              <button className="rounded bg-slate-700 px-3 py-2 text-white" onClick={() => handleAuth('login')}>Login</button>
            </div>
          </>
        ) : (
          <>
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">Submit idea</h2>
              <button className="rounded border px-3 py-1" onClick={logout}>Logout ({email})</button>
            </div>
            <div className="mt-3 grid gap-2">
              <input className="rounded border p-2" placeholder="Idea title" value={title} onChange={(e) => setTitle(e.target.value)} />
              <textarea className="min-h-28 rounded border p-2" placeholder="Idea description" value={description} onChange={(e) => setDescription(e.target.value)} />
              <button className="w-fit rounded bg-blue-600 px-3 py-2 text-white" onClick={submitIdea}>Create idea</button>
            </div>
          </>
        )}
      </section>

      <section className="mt-6 rounded-xl bg-white p-4 shadow">
        <h2 className="text-lg font-semibold">Ideas</h2>
        <ul className="mt-3 space-y-3">
          {ideas.map((idea) => (
            <li key={idea.id} className="rounded border p-3">
              <div className="font-semibold">{idea.title}</div>
              <p className="text-slate-700">{idea.description}</p>
              <p className="text-sm text-slate-500">by {idea.userEmail} · {new Date(idea.createdAt).toLocaleString()}</p>
            </li>
          ))}
        </ul>
      </section>
    </main>
  )
}

createRoot(document.getElementById('root')!).render(<App />)
