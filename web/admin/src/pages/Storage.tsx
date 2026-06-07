import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { 
  Upload, 
  Trash2, 
  Folder, 
  File, 
  Plus
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { storageApi } from '@/lib/api'

export function StoragePage() {
  const queryClient = useQueryClient()
  const [selectedBucket, setSelectedBucket] = useState<string | null>(null)
  const [newBucketName, setNewBucketName] = useState('')
  const [uploadFile, setUploadFile] = useState<File | null>(null)

  const { data: bucketsData, isLoading: bucketsLoading } = useQuery({
    queryKey: ['buckets'],
    queryFn: async () => {
      const res = await storageApi.listBuckets()
      return res.data
    },
  })

  const { data: objectsData, isLoading: objectsLoading } = useQuery({
    queryKey: ['objects', selectedBucket],
    queryFn: async () => {
      if (!selectedBucket) return { objects: [] }
      const res = await storageApi.listObjects(selectedBucket)
      return res.data
    },
    enabled: !!selectedBucket,
  })

  const createBucketMutation = useMutation({
    mutationFn: (name: string) => storageApi.createBucket(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['buckets'] })
      setNewBucketName('')
    },
  })

  const deleteBucketMutation = useMutation({
    mutationFn: (name: string) => storageApi.deleteBucket(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['buckets'] })
      setSelectedBucket(null)
    },
  })

  const uploadMutation = useMutation({
    mutationFn: ({ bucket, path, file }: { bucket: string; path: string; file: File }) =>
      storageApi.upload(bucket, path, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', selectedBucket] })
      setUploadFile(null)
    },
  })

  const deleteObjectMutation = useMutation({
    mutationFn: ({ bucket, path }: { bucket: string; path: string }) =>
      storageApi.deleteObject(bucket, path),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', selectedBucket] })
    },
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Storage</h1>
        <Dialog>
          <DialogTrigger asChild>
            <Button><Plus className="mr-2 h-4 w-4" /> New Bucket</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Bucket</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <Input
                placeholder="Bucket name"
                value={newBucketName}
                onChange={(e) => setNewBucketName(e.target.value)}
              />
              <Button 
                onClick={() => createBucketMutation.mutate(newBucketName)}
                disabled={!newBucketName || createBucketMutation.isPending}
              >
                Create
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="md:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Folder className="h-5 w-5" /> Buckets
            </CardTitle>
          </CardHeader>
          <CardContent>
            {bucketsLoading ? (
              <div className="text-sm text-muted-foreground">Loading...</div>
            ) : (
              <div className="space-y-2">
                {bucketsData?.buckets.map((bucket: any) => (
                  <div
                    key={bucket.name}
                    className={`flex items-center justify-between p-2 rounded cursor-pointer ${
                      selectedBucket === bucket.name ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
                    }`}
                    onClick={() => setSelectedBucket(bucket.name)}
                  >
                    <span className="text-sm font-medium">{bucket.name}</span>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={(e) => {
                        e.stopPropagation()
                        deleteBucketMutation.mutate(bucket.name)
                      }}
                    >
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span className="flex items-center gap-2">
                <File className="h-5 w-5" /> 
                {selectedBucket ? `Objects in ${selectedBucket}` : 'Select a bucket'}
              </span>
              {selectedBucket && (
                <div className="flex items-center gap-2">
                  <Input
                    type="file"
                    className="w-auto"
                    onChange={(e) => setUploadFile(e.target.files?.[0] || null)}
                  />
                  <Button
                    size="sm"
                    disabled={!uploadFile}
                    onClick={() => {
                      if (uploadFile && selectedBucket) {
                        uploadMutation.mutate({
                          bucket: selectedBucket,
                          path: uploadFile.name,
                          file: uploadFile,
                        })
                      }
                    }}
                  >
                    <Upload className="mr-2 h-4 w-4" /> Upload
                  </Button>
                </div>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {!selectedBucket ? (
              <div className="text-sm text-muted-foreground">Select a bucket to view objects</div>
            ) : objectsLoading ? (
              <div className="text-sm text-muted-foreground">Loading...</div>
            ) : (
              <div className="space-y-2">
                {objectsData?.objects.map((obj: any) => (
                  <div key={obj.name} className="flex items-center justify-between p-2 rounded hover:bg-muted">
                    <div className="flex items-center gap-2">
                      <File className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm">{obj.name}</span>
                      <span className="text-xs text-muted-foreground">
                        {(obj.size / 1024).toFixed(1)} KB
                      </span>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={() => deleteObjectMutation.mutate({ bucket: selectedBucket, path: obj.name })}
                    >
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
