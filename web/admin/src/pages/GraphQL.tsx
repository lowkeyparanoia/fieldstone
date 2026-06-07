import { useState } from 'react'
import { GitGraph, Copy, Check } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { graphqlApi } from '@/lib/api'

export function GraphQLPage() {
  const [query, setQuery] = useState(`query {
  users {
    id
    email
    created_at
  }
}`)
  const [variables, setVariables] = useState('{}')
  const [result, setResult] = useState('')
  const [copied, setCopied] = useState(false)

  const executeQuery = async () => {
    try {
      const vars = variables ? JSON.parse(variables) : {}
      const res = await graphqlApi.query(query, vars)
      setResult(JSON.stringify(res.data, null, 2))
    } catch (err: any) {
      setResult(JSON.stringify({ error: err.message }, null, 2))
    }
  }

  const copyResult = () => {
    navigator.clipboard.writeText(result)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <GitGraph className="h-8 w-8" /> GraphQL Playground
        </h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Query</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Textarea
              value={query}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setQuery(e.target.value)}
              className="font-mono min-h-[300px]"
              placeholder="Enter GraphQL query..."
            />
            <Textarea
              value={variables}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setVariables(e.target.value)}
              className="font-mono min-h-[80px]"
              placeholder='Variables: { "key": "value" }'
            />
            <Button onClick={executeQuery} className="w-full">
              Execute
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Result</CardTitle>
            {result && (
              <Button variant="ghost" size="sm" onClick={copyResult}>
                {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
              </Button>
            )}
          </CardHeader>
          <CardContent>
            <pre className="font-mono text-sm bg-muted p-4 rounded-md min-h-[400px] overflow-auto">
              {result || 'Execute a query to see results'}
            </pre>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
