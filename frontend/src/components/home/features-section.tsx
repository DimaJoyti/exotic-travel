import { Shield, Globe, Clock, Award, Users, Headphones } from 'lucide-react'

const features = [
  {
    icon: Globe,
    title: 'Exotic Destinations',
    description: 'Discover unique and breathtaking locations that most travelers never experience.',
  },
  {
    icon: Shield,
    title: 'Secure Booking',
    description: 'Your payments and personal information are protected with bank-level security.',
  },
  {
    icon: Clock,
    title: '24/7 Support',
    description: 'Our travel experts are available around the clock to assist with your journey.',
  },
  {
    icon: Award,
    title: 'Best Price Guarantee',
    description: 'We guarantee the best prices for all our destinations with no hidden fees.',
  },
  {
    icon: Users,
    title: 'Expert Guides',
    description: 'Local experts and certified guides ensure authentic and safe experiences.',
  },
  {
    icon: Headphones,
    title: 'Personalized Service',
    description: 'Tailored itineraries and personalized recommendations for your perfect trip.',
  },
]

export default function FeaturesSection() {
  return (
    <section className="py-16 bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="text-center mb-12">
          <h2 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
            Why Choose ExoticTravel?
          </h2>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto">
            We're committed to providing you with extraordinary travel experiences 
            that create memories to last a lifetime
          </p>
        </div>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => {
            const IconComponent = feature.icon
            return (
              <div
                key={index}
                className="text-center p-6 rounded-xl hover:bg-gray-50 transition-colors duration-200 group"
              >
                <div className="inline-flex items-center justify-center w-16 h-16 bg-primary/10 rounded-full mb-4 group-hover:bg-primary/20 transition-colors">
                  <IconComponent className="h-8 w-8 text-primary" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-3">
                  {feature.title}
                </h3>
                <p className="text-gray-600 leading-relaxed">
                  {feature.description}
                </p>
              </div>
            )
          })}
        </div>

        {/* Stats Section */}
        <div className="mt-16 bg-gradient-to-r from-primary to-primary/80 rounded-2xl p-8 text-white">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
            <div>
              <div className="text-3xl md:text-4xl font-bold mb-2">50+</div>
              <div className="text-primary-foreground/80">Destinations</div>
            </div>
            <div>
              <div className="text-3xl md:text-4xl font-bold mb-2">10K+</div>
              <div className="text-primary-foreground/80">Happy Travelers</div>
            </div>
            <div>
              <div className="text-3xl md:text-4xl font-bold mb-2">98%</div>
              <div className="text-primary-foreground/80">Satisfaction Rate</div>
            </div>
            <div>
              <div className="text-3xl md:text-4xl font-bold mb-2">24/7</div>
              <div className="text-primary-foreground/80">Support</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
