export default function FeaturesSection() {
  return (
    <section className="py-20 bg-gradient-to-b from-white to-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-6">
            Why Choose ExoticTravel?
          </h2>
          <p className="text-xl text-gray-600 max-w-4xl mx-auto leading-relaxed">
            We're committed to providing you with extraordinary travel experiences
            that create memories to last a lifetime
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div className="text-center p-8 rounded-2xl bg-white shadow-lg border border-gray-100">
            <div className="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-br from-blue-100 to-blue-200 rounded-2xl mb-6">
              <div className="h-10 w-10 bg-blue-600 rounded"></div>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mb-4">
              Exotic Destinations
            </h3>
            <p className="text-gray-600 leading-relaxed text-lg">
              Discover unique and breathtaking locations that most travelers never experience.
            </p>
          </div>

          <div className="text-center p-8 rounded-2xl bg-white shadow-lg border border-gray-100">
            <div className="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-br from-blue-100 to-blue-200 rounded-2xl mb-6">
              <div className="h-10 w-10 bg-blue-600 rounded"></div>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mb-4">
              Secure Booking
            </h3>
            <p className="text-gray-600 leading-relaxed text-lg">
              Your payments and personal information are protected with bank-level security.
            </p>
          </div>

          <div className="text-center p-8 rounded-2xl bg-white shadow-lg border border-gray-100">
            <div className="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-br from-blue-100 to-blue-200 rounded-2xl mb-6">
              <div className="h-10 w-10 bg-blue-600 rounded"></div>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mb-4">
              24/7 Support
            </h3>
            <p className="text-gray-600 leading-relaxed text-lg">
              Our travel experts are available around the clock to assist with your journey.
            </p>
          </div>
        </div>

        <div className="mt-20 bg-gradient-to-r from-blue-500 via-purple-500 to-blue-600 rounded-3xl p-12 text-white shadow-2xl">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
            <div>
              <div className="text-4xl md:text-5xl font-bold mb-3">50+</div>
              <div className="text-white/90 text-lg font-medium">Destinations</div>
            </div>
            <div>
              <div className="text-4xl md:text-5xl font-bold mb-3">10K+</div>
              <div className="text-white/90 text-lg font-medium">Happy Travelers</div>
            </div>
            <div>
              <div className="text-4xl md:text-5xl font-bold mb-3">98%</div>
              <div className="text-white/90 text-lg font-medium">Satisfaction Rate</div>
            </div>
            <div>
              <div className="text-4xl md:text-5xl font-bold mb-3">24/7</div>
              <div className="text-white/90 text-lg font-medium">Support</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
