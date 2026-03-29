'use client';

import { useEffect, useRef } from 'react';
import { motion, useAnimation } from 'framer-motion';
import React from 'react';
import LiveMapPreview from './LiveMapPreview';

const fadeInUp = {
  initial: { opacity: 0, y: 30 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.6 }
};

const InteractiveIntelligence = () => {
  const sectionRef = useRef<HTMLElement>(null);
  const mapRef = useRef<HTMLDivElement>(null);
  const textRef = useRef<HTMLDivElement>(null);
  const controls = useAnimation();

  useEffect(() => {
    const gsap = (window as any).gsap;
    const ScrollTrigger = (window as any).ScrollTrigger;
    if (!gsap) return;

    const ctx = gsap.context(() => {
      // Pin the map container while text scrolls
      if (ScrollTrigger) {
        ScrollTrigger.create({
          trigger: sectionRef.current,
          start: 'top top',
          end: 'bottom top',
          pin: mapRef.current,
          pinSpacing: true,
          scrub: 1
        });

        // Parallax: background depth moves slower
        gsap.to(mapRef.current, {
          backgroundPositionY: '20%',
          ease: 'none',
          scrollTrigger: {
            trigger: sectionRef.current,
            start: 'top bottom',
            end: 'bottom top',
            scrub: true
          }
        });

        // Text slides over map as you scroll
        gsap.to(textRef.current, {
          y: -100,
          opacity: 0.5,
          ease: 'none',
          scrollTrigger: {
            trigger: sectionRef.current,
            start: 'top 40%',
            end: 'bottom 10%',
            scrub: 1
          }
        });

        // Animate efficiency bar based on scroll
        gsap.to('#efficiency-bar', {
          width: '94%',
          ease: 'none',
          scrollTrigger: {
            trigger: sectionRef.current,
            start: 'top 30%',
            end: 'bottom 40%',
            scrub: 1
          }
        });

        // Count-up animation for vehicle count
        const vehicleCount = document.getElementById('vehicle-count');
        if (vehicleCount) {
          ScrollTrigger.create({
            trigger: sectionRef.current,
            start: 'top 50%',
            end: 'bottom 30%',
            onEnter: () => {
              gsap.to(vehicleCount, {
                innerText: 128,
                duration: 2,
                snap: { innerText: 1 },
                ease: 'power1.out'
              });
            },
            onLeaveBack: () => {
              gsap.set(vehicleCount, { innerText: 0 });
            }
          });
        }

        // Count-up for alerts
        const alertsCount = document.getElementById('alerts-count');
        if (alertsCount) {
          ScrollTrigger.create({
            trigger: sectionRef.current,
            start: 'top 60%',
            end: 'bottom 20%',
            onEnter: () => {
              gsap.to(alertsCount, {
                innerText: 2,
                duration: 1,
                snap: { innerText: 1 },
                ease: 'power1.out'
              });
            },
            onLeaveBack: () => {
              gsap.set(alertsCount, { innerText: 0 });
            }
          });
        }

        // Feature cards fade in sequentially as you scroll
        gsap.utils.toArray('.feature-card').forEach((card: Element, i: number) => {
          gsap.from(card, {
            opacity: 0,
            x: i === 0 ? -50 : 50,
            duration: 0.6,
            scrollTrigger: {
              trigger: sectionRef.current,
              start: `top ${65 + i * 15}%`,
              end: `bottom ${55 + i * 15}%`,
              toggleActions: 'play none none reverse',
              scrub: 1
            }
          });
        });
      }
    }, sectionRef);

    return () => ctx.revert();
  }, []);

  return (
    <section
      ref={sectionRef}
      className="scene relative"
      style={{ '--scene-height': '300vh', paddingTop: '0', paddingBottom: '0' } as React.CSSProperties}
      data-scene="intelligence"
    >
      {/* Background layer - pinned map */}
      <div
        ref={mapRef}
        className="absolute inset-0 z-0 px-4 md:px-10 pt-24 pb-24 flex items-center justify-center"
      >
        <LiveMapPreview />
      </div>

      {/* Text content - scrolls over pinned map */}
      <div
        ref={textRef}
        className="relative z-10 max-w-[1280px] mx-auto pt-32"
      >
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-16 items-center">
          <div className="lg:col-span-5">
            <motion.div {...fadeInUp}>
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-black uppercase tracking-widest mb-6">
                Live Intelligence
              </div>
              <h2 className="text-4xl md:text-5xl font-black text-slate-900 dark:text-white mb-8 font-display leading-[1.1]">
                See Your Fleet <br/>
                <span className="text-primary italic">In Motion.</span>
              </h2>
              <p className="text-slate-600 dark:text-text-secondary text-lg mb-10 leading-relaxed">
                Stop guessing and start knowing. Our real-time engine processes billions of data points to give you a crystal-clear view of every vehicle, every route, and every cost-saving opportunity.
              </p>
              <div className="space-y-6">
                <div className="feature-card flex gap-4 p-4 rounded-2xl bg-white/90 dark:bg-surface-dark/90 backdrop-blur-xl border border-slate-200 dark:border-border-dark shadow-lg">
                  <div className="size-10 rounded-xl bg-primary/10 flex items-center justify-center">
                    <span className="material-symbols-outlined !text-xl text-primary">bolt</span>
                  </div>
                  <div>
                    <h4 className="font-bold text-slate-900 dark:text-white">5-Second Latency</h4>
                    <p className="text-sm text-slate-500">The fastest data refresh rates in the industry.</p>
                  </div>
                </div>
                <div className="feature-card flex gap-4 p-4 rounded-2xl bg-white/90 dark:bg-surface-dark/90 backdrop-blur-xl border border-slate-200 dark:border-border-dark shadow-lg">
                  <div className="size-10 rounded-xl bg-primary/10 flex items-center justify-center">
                    <span className="material-symbols-outlined !text-xl text-primary">auto_graph</span>
                  </div>
                  <div>
                    <h4 className="font-bold text-slate-900 dark:text-white">Predictive ETAs</h4>
                    <p className="text-sm text-slate-500">AI-calculated arrivals based on live traffic and weather.</p>
                  </div>
                </div>
              </div>
            </motion.div>
          </div>
          <div className="lg:col-span-7"></div> {/* Empty column for layout balance */}
        </div>
      </div>
    </section>
  );
};

export default InteractiveIntelligence;
