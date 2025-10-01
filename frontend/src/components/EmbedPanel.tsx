"use client"

import type React from "react"
import { useState, useEffect } from "react"
import FileUpload from "./FileUpload"
import AudioPlayer from "./AudioPlayer"
import Button from "./Button"
import type { UploadedFile, EmbedOptions, EmbedResult, AppStatus, CapacityInfo } from "../types"
import StatusDisplay from "./StatusDisplay"
import { STEGANOGRAPHY_METHODS, DEFAULT_EMBED_OPTIONS, getCapacityForMethod, formatCapacity } from "../utils/steganography"

interface EmbedPanelProps {}

const EmbedPanel: React.FC<EmbedPanelProps> = () => {
  const [coverAudio, setCoverAudio] = useState<UploadedFile | undefined>()
  const [stegoAudio, setStegoAudio] = useState<UploadedFile | undefined>()
  const [secretFile, setSecretFile] = useState<UploadedFile | undefined>()
  const [options, setOptions] = useState<EmbedOptions>(DEFAULT_EMBED_OPTIONS)
  const [isEmbedding, setIsEmbedding] = useState(false)

  const API_URL = "http://localhost:8080"

  const [status, setStatus] = useState<AppStatus>({
    isLoading: false,
    message: "CYBERSTEG TERMINAL READY - SELECT OPERATION MODE",
    type: "info",
  })
  const [capacities, setCapacities] = useState<{ [key: string]: number }>({})
  const [psnr, setPsnr] = useState<number | undefined>()

  const handleStatusUpdate = (newStatus: AppStatus) => {
    setStatus(newStatus)
  }

  const handleEmbedComplete = (result: EmbedResult) => {
    if (result.success && result.psnr) {
      setPsnr(result.psnr)
      setStatus({
        isLoading: false,
        message: `MESSAGE EMBEDDED SUCCESSFULLY! PSNR: ${result.psnr.toFixed(2)} dB`,
        type: "success",
      })
    } else {
      setStatus({
        isLoading: false,
        message: result.message || "EMBEDDING OPERATION FAILED",
        type: "error",
      })
    }
  }

  const handleCoverAudioSelect = async (file: UploadedFile) => {
    setCoverAudio(file)
    setStegoAudio(undefined) // Clear previous result
    setCapacities({}) // Reset capacities

    if (file) {
      handleStatusUpdate({
        isLoading: true,
        message: "Analyzing audio capacity...",
        type: "info",
      })

      try {
        const formData = new FormData()
        formData.append("audio", file.file)

        const response = await fetch(`${API_URL}/api/v1/capacity`, {
          method: "POST",
          body: formData,
        })

        if (!response.ok) {
          const errorData = await response.json()
          throw new Error(errorData.message || "Failed to calculate capacity")
        }

        const data = await response.json()
        setCapacities(data.capacities)

        handleStatusUpdate({
          isLoading: false,
          message: "Audio capacity analyzed successfully",
          type: "success",
        })
      } catch (error) {
        handleStatusUpdate({
          isLoading: false,
          message: error instanceof Error ? error.message : "Failed to analyze capacity",
          type: "error",
        })
      }
    }
  }

  const handleSecretFileSelect = (file: UploadedFile) => {
    setSecretFile(file)
  }

  const handleEmbed = async () => {
    if (!coverAudio || !secretFile) return

    setIsEmbedding(true)
    handleStatusUpdate({
      isLoading: true,
      message: "Embedding file into audio...",
      type: "info",
    })

    try {
      const formData = new FormData()
      formData.append("audio", coverAudio.file)
      formData.append("secret", secretFile.file)
      formData.append("method", options.method)
      if (options.method === 'lsb') {
        formData.append("lsb", options.nLSB.toString())
      }
      formData.append("use_encryption", options.encrypt.toString())
      formData.append("use_random_start", options.randomStart.toString())
      if (options.stegKey) {
        formData.append("stego_key", options.stegKey)
      }
      formData.append("output_filename", `stego_${coverAudio.name}`)

      const response = await fetch(`${API_URL}/api/v1/embed`, {
        method: "POST",
        body: formData,
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || "Failed to embed file")
      }

      const contentDisposition = response.headers.get("Content-Disposition")
      const filenameMatch = contentDisposition?.match(/filename="([^"]+)"/)
      const filename = filenameMatch ? filenameMatch[1] : `stego_${coverAudio.name}`

      const blob = await response.blob()
      const stegoFile: UploadedFile = {
        file: new File([blob], filename, { type: "audio/mpeg" }),
        name: filename,
        size: blob.size,
        url: URL.createObjectURL(blob),
      }

      setStegoAudio(stegoFile)

      const result: EmbedResult = {
        success: true,
        psnr: parseFloat(response.headers.get("X-PSNR-Value") || "0"),
        stegoAudioUrl: stegoFile.url,
        message: "File successfully embedded",
        secretSize: parseInt(response.headers.get("X-Secret-Size") || "0"),
        processingTime: parseInt(response.headers.get("X-Processing-Time") || "0"),
        embeddingMethod: response.headers.get("X-Embedding-Method") || "LSB",
      }

      handleEmbedComplete(result)
      handleStatusUpdate({
        isLoading: false,
        message: "Embedding completed successfully",
        type: "success",
      })
    } catch (error) {
      const result: EmbedResult = {
        success: false,
        message: error instanceof Error ? error.message : "Failed to embed file",
      }
      handleEmbedComplete(result)
      handleStatusUpdate({
        isLoading: false,
        message: error instanceof Error ? error.message : "Failed to embed file",
        type: "error",
      })
    } finally {
      setIsEmbedding(false)
    }
  }

  const handleDownload = () => {
    if (!stegoAudio) return

    const link = document.createElement("a")
    link.href = stegoAudio.url
    link.download = stegoAudio.name
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  // Check capacity sufficiency when relevant states change
  useEffect(() => {
    if (coverAudio && secretFile && Object.keys(capacities).length > 0) {
      const selectedCapacity = getCapacityForMethod(capacities, options.method, options.nLSB)
      const methodName = STEGANOGRAPHY_METHODS[options.method].name
      
      if (secretFile.size > selectedCapacity) {
        handleStatusUpdate({
          isLoading: false,
          message: `Secret file too large for ${methodName}. Max capacity: ${formatCapacity(selectedCapacity)}`,
          type: "error",
        })
      } else {
        handleStatusUpdate({
          isLoading: false,
          message: `Capacity sufficient for embedding with ${methodName}`,
          type: "success",
        })
      }
    }
  }, [coverAudio, secretFile, options.method, options.nLSB, capacities])

  const isCapacitySufficient = () => {
    if (!secretFile || Object.keys(capacities).length === 0) return true
    const selectedCapacity = getCapacityForMethod(capacities, options.method, options.nLSB)
    return secretFile.size <= selectedCapacity
  }

  const getCapacityInfo = (): CapacityInfo | undefined => {
    if (!secretFile || Object.keys(capacities).length === 0) return undefined
    
    return {
      audioCapacity: capacities as any,
      selectedMethod: options.method,
      selectedLSB: options.method === 'lsb' ? options.nLSB : undefined,
      isCapacitySufficient: isCapacitySufficient(),
      requiredBytes: secretFile.size,
      availableBytes: getCapacityForMethod(capacities, options.method, options.nLSB)
    }
  }

  const isEmbedReady = coverAudio && secretFile && (options.encrypt || options.randomStart ? options.stegKey.trim() : true) && isCapacitySufficient()

  const getKeyPlaceholder = () => {
    if (options.encrypt && options.randomStart) {
      return "Enter secret key for encryption and random seed..."
    } else if (options.encrypt) {
      return "Enter secret key for encryption..."
    } else if (options.randomStart) {
      return "Enter secret key for random seed..."
    }
    return ""
  }

  return (
    <div className="space-y-8">
      {/* Cover Audio Upload */}
      <div className="border border-purple-400/50 rounded-lg p-6 bg-purple-900/10 backdrop-blur-sm">
        <FileUpload
          onFileSelect={handleCoverAudioSelect}
          accept="audio/*"
          label="COVER AUDIO FILE"
          currentFile={coverAudio}
        />
      </div>

      {/* Secret File Upload */}
      <div className="border border-cyan-400/50 rounded-lg p-6 bg-black/30 backdrop-blur-sm">
        <FileUpload
          onFileSelect={handleSecretFileSelect}
          accept="*/*"
          label="SECRET FILE"
          currentFile={secretFile}
        />
      </div>

      {/* Advanced Options */}
      <div className="border border-yellow-400/50 rounded-lg p-6 bg-yellow-900/10 backdrop-blur-sm">
        <div className="text-yellow-400 font-mono text-sm mb-4 flex items-center gap-2">
          <span className="w-2 h-2 bg-yellow-400 rounded-full animate-pulse"></span>
          ADVANCED OPTIONS
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={options.encrypt}
              onChange={(e) => setOptions({ ...options, encrypt: e.target.checked })}
              className="w-5 h-5 text-yellow-400 bg-black border-yellow-400 rounded focus:ring-yellow-400 focus:ring-2"
            />
            <span className="text-white font-mono">VIGENERE CIPHER ENCRYPTION</span>
          </label>

          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={options.randomStart}
              onChange={(e) => setOptions({ ...options, randomStart: e.target.checked })}
              className="w-5 h-5 text-yellow-400 bg-black border-yellow-400 rounded focus:ring-yellow-400 focus:ring-2"
            />
            <span className="text-white font-mono">RANDOM START POSITION</span>
          </label>
        </div>

        {(options.encrypt || options.randomStart) && (
          <div className="mt-6 border border-pink-500/50 rounded-lg p-6 bg-pink-900/10 backdrop-blur-sm">
            <div className="text-pink-400 font-mono text-sm mb-4 flex items-center gap-2">
              <span className="w-2 h-2 bg-pink-400 rounded-full animate-pulse"></span>
              SECRET KEY
            </div>

            <input
              type="text"
              value={options.stegKey}
              onChange={(e) => setOptions({ ...options, stegKey: e.target.value })}
              placeholder={getKeyPlaceholder()}
              className="w-full bg-black/50 border border-pink-400/30 rounded-lg p-3 text-white font-mono focus:border-pink-400 focus:outline-none focus:ring-2 focus:ring-pink-400/25 transition-all"
            />
          </div>
        )}
      </div>

      {/* Method Selection */}
      <div className="border border-green-400/50 rounded-lg p-6 bg-green-900/10 backdrop-blur-sm">
        <div className="text-green-400 font-mono text-sm mb-4 flex items-center gap-2">
          <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
          STEGANOGRAPHY METHOD
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
          {Object.values(STEGANOGRAPHY_METHODS).map((method) => (
            <div
              key={method.id}
              className={`border rounded-lg p-4 cursor-pointer transition-all ${
                options.method === method.id
                  ? 'border-green-400 bg-green-400/10'
                  : 'border-green-400/30 bg-black/30 hover:border-green-400/50'
              }`}
              onClick={() => setOptions({ ...options, method: method.id })}
            >
              <div className="flex items-center gap-3 mb-2">
                <input
                  type="radio"
                  name="method"
                  value={method.id}
                  checked={options.method === method.id}
                  onChange={() => {}} // Handled by parent div onClick
                  className="w-4 h-4 text-green-400 bg-black border-green-400 focus:ring-green-400 focus:ring-2"
                />
                <span className="text-white font-mono font-bold">{method.name}</span>
                <div className="flex gap-1">
                  <span className={`px-2 py-0.5 text-xs rounded ${
                    method.capacity === 'high' ? 'bg-green-500/20 text-green-400' :
                    method.capacity === 'medium' ? 'bg-yellow-500/20 text-yellow-400' :
                    'bg-red-500/20 text-red-400'
                  }`}>
                    {method.capacity} capacity
                  </span>
                  <span className={`px-2 py-0.5 text-xs rounded ${
                    method.robustness === 'high' ? 'bg-green-500/20 text-green-400' :
                    method.robustness === 'medium' ? 'bg-yellow-500/20 text-yellow-400' :
                    'bg-red-500/20 text-red-400'
                  }`}>
                    {method.robustness} robustness
                  </span>
                </div>
              </div>
              <p className="text-gray-300 text-sm mb-3">{method.description}</p>
              
              <div className="space-y-2">
                <div>
                  <p className="text-green-300 text-xs font-semibold mb-1">Advantages:</p>
                  <ul className="text-xs text-gray-400 space-y-0.5">
                    {method.advantages.slice(0, 2).map((advantage, idx) => (
                      <li key={idx}>• {advantage}</li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* LSB Configuration (only show for LSB method) */}
        {options.method === 'lsb' && (
          <div className="border border-green-400/30 rounded-lg p-4 bg-black/20">
            <label className="text-white font-mono text-sm mb-3 block">LSB BITS CONFIGURATION:</label>
            <div className="flex gap-4">
              {[1, 2, 3, 4].map((bits) => (
                <label key={bits} className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="nLSB"
                    value={bits}
                    checked={options.nLSB === bits}
                    onChange={(e) => setOptions({ ...options, nLSB: Number.parseInt(e.target.value) as 1 | 2 | 3 | 4 })}
                    className="w-4 h-4 text-green-400 bg-black border-green-400 focus:ring-green-400 focus:ring-2"
                  />
                  <span className="text-white font-mono">{bits}</span>
                </label>
              ))}
            </div>
            <p className="text-gray-400 text-xs mt-2">
              Higher values = more capacity but potentially lower audio quality
            </p>
          </div>
        )}

        {/* Method-specific capacity info */}
        {secretFile && Object.keys(capacities).length > 0 && (
          <div className="mt-4 p-3 bg-black/40 border border-green-400/20 rounded">
            <div className="text-green-400 text-xs font-mono mb-2">CAPACITY ANALYSIS</div>
            <div className="grid grid-cols-2 gap-4 text-xs">
              <div>
                <span className="text-gray-400">Required:</span>
                <span className="text-white ml-2">{formatCapacity(secretFile.size)}</span>
              </div>
              <div>
                <span className="text-gray-400">Available:</span>
                <span className={`ml-2 ${isCapacitySufficient() ? 'text-green-400' : 'text-red-400'}`}>
                  {formatCapacity(getCapacityForMethod(capacities, options.method, options.nLSB))}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>

      <StatusDisplay status={status} capacityInfo={getCapacityInfo()} psnr={psnr} />

      {/* Embed Button */}
      <div className="flex justify-center">
        <Button
          onClick={handleEmbed}
          disabled={!isEmbedReady || isEmbedding}
          variant="primary"
          className="px-12 py-4 text-lg"
        >
          {isEmbedding ? "EMBEDDING..." : "EMBED FILE"}
        </Button>
      </div>

      {/* Audio Players */}
      {coverAudio && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <AudioPlayer audioUrl={coverAudio.url} label="ORIGINAL AUDIO" />
          {stegoAudio && (
            <div>
              <AudioPlayer audioUrl={stegoAudio.url} label="STEGO AUDIO" onClick={handleDownload} />
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default EmbedPanel
