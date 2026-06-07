import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { FunctionSquare, Send } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { functionsApi } from '@/lib/api'

const FUNCTIONS = [
  'video-token',
  'score-compute',
  'fitbit-sync',
  'fitbit-oauth',
  'booking-noshow',
  'sms-send',
  'razorpay-webhook',
  'account-delete',
  'enrich',
]

export function FunctionsPage() {
  const [selectedFunction, setSelectedFunction] = useState(FUNCTIONS[0])
  const [payload, setPayload] = useState('{}')
  const [result, setResult] = useState('')

  const invokeMutation = useMutation({
    mutationFn: () => functionsApi.invoke(selectedFunction, JSON.parse(payload)),
    onSuccess: (res) => {
      setResult(JSON.stringify(res.data, null, 2))
    },
    onError: (err: any) => {
      setResult(JSON.stringify({ error: err.message }, null, 2))
    },
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <FunctionSquare className="h-8 w-8" /> Edge Functions
        </h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Functions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {FUNCTIONS.map((fn) => (
                <Button
                  key={fn}
                  variant={selectedFunction === fn ? 'default' : 'ghost'}
                  className="w-full justify-start"
                  onClick={() => setSelectedFunction(fn)}
                >
                  <FunctionSquare className="mr-2 h-4 w-4" />
                  {fn}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Invoke: {selectedFunction}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Textarea
              value={payload}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setPayload(e.target.value)}
              className="font-mono min-h-[150px]"
              placeholder='Payload: { "key": "value" }'
            />
            <Button 
              onClick={() => invokeMutation.mutate()}
              disabled={invokeMutation.isPending}
              className="w-full"
            >
              <Send className="mr-2 h-4 w-4" />
              {invokeMutation.isPending ? 'Invoking...' : 'Invoke'}
            </Button>
            {result && (
              <pre className="font-mono text-sm bg-muted p-4 rounded-md overflow-auto">
                {result}
              </pre>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
