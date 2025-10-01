import type React from "react"

const Header: React.FC = () => {
  return (
    <header className="text-center mb-12">
      <div className="relative">
        {/* Main title with CRT effect */}
        <h1 className="text-6xl md:text-8xl font-black mb-4 relative">
          <span className="bg-gradient-to-r from-pink-500 via-purple-500 to-cyan-400 bg-clip-text text-transparent neon-text">
            CYBERSTEG
          </span>
          {/* CRT glow effect */}
          <div className="absolute inset-0 bg-gradient-to-r from-pink-500/20 via-purple-500/20 to-cyan-400/20 blur-xl -z-10"></div>
        </h1>

        {/* Subtitle */}
        <p className="text-xl md:text-2xl text-cyan-400 font-light tracking-wider mb-2">AUDIO STEGANOGRAPHY TERMINAL</p>
      </div>

      {/* Grid pattern background */}
      <div className="absolute inset-0 opacity-10 pointer-events-none">
        <div className="grid grid-cols-12 gap-1 h-full">
          {Array.from({ length: 144 }).map((_, i) => (
            <div key={i} className="border border-cyan-400/20"></div>
          ))}
        </div>
      </div>
    </header>
  )
}

export default Header
