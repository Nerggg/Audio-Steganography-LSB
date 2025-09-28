"use client"

import type React from "react"
import { useState, useEffect } from "react"
import FileUpload from "./FileUpload"
import AudioPlayer from "./AudioPlayer"
import Button from "./Button"
import type { UploadedFile, EmbedOptions, EmbedResult, AppStatus } from "../types"
import StatusDisplay from "./StatusDisplay"

interface EmbedPanelProps {}

const EmbedPanel: React.FC<EmbedPanelProps> = () => {
  const [coverAudio, setCoverAudio] = useState<UploadedFile | undefined>()
  const [stegoAudio, setStegoAudio] = useState<UploadedFile | undefined>()
  const [secretFile, setSecretFile] = useState<UploadedFile | undefined>()
  const [options, setOptions] = useState<EmbedOptions>({
    stegKey: "",
    nLSB: 1,
    encrypt: false,
    randomStart: false,
  })
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
      formData.append("lsb", options.nLSB.toString())
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
      const selectedCapacity = capacities[`${options.nLSB}_lsb`] || 0
      if (secretFile.size > selectedCapacity) {
        handleStatusUpdate({
          isLoading: false,
          message: `Secret file too large for selected LSB configuration. Max capacity: ${selectedCapacity} bytes`,
          type: "error",
        })
      } else {
        handleStatusUpdate({
          isLoading: false,
          message: "Capacity sufficient for embedding",
          type: "success",
        })
      }
    }
  }, [coverAudio, secretFile, options.nLSB, capacities])

  const isCapacitySufficient = () => {
    if (!secretFile || Object.keys(capacities).length === 0) return true
    const selectedCapacity = capacities[`${options.nLSB}_lsb`] || 0
    return secretFile.size <= selectedCapacity
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

      {/* LSB Options */}
      <div className="border border-green-400/50 rounded-lg p-6 bg-green-900/10 backdrop-blur-sm">
        <div className="text-green-400 font-mono text-sm mb-4 flex items-center gap-2">
          <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
          LSB CONFIGURATION
        </div>

        <div className="space-y-3">
          <div>
            <label className="text-white font-mono text-sm mb-2 block">N-LSB BITS:</label>
            <div className="flex gap-4">
              {[1, 2, 4].map((bits) => (
                <label key={bits} className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="nLSB"
                    value={bits}
                    checked={options.nLSB === bits}
                    onChange={(e) => setOptions({ ...options, nLSB: Number.parseInt(e.target.value) as 1 | 2 | 4 })}
                    className="w-4 h-4 text-green-400 bg-black border-green-400 focus:ring-green-400 focus:ring-2"
                  />
                  <span className="text-white font-mono">{bits}</span>
                </label>
              ))}
            </div>
          </div>
        </div>
      </div>

      <StatusDisplay status={status} psnr={psnr} />

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
