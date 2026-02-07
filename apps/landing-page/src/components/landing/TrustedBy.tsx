'use client';


const TrustedBy = () => {
  return (
    <section className="border-y border-slate-200 dark:border-border-dark bg-slate-50 dark:bg-background-dark/50 py-10">
      <div className="max-w-[1280px] mx-auto px-4 md:px-10">
        <p className="text-center text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-8">Trusted by Global Logistics Leaders</p>
        <div className="flex flex-wrap justify-center items-center gap-12 md:gap-20 opacity-60 grayscale hover:grayscale-0 transition-all duration-500">
          {['ATLANTIC:anchor', 'AIRCARGO:flight_takeoff', 'TRANSROAD:local_shipping', 'BOXLOGISTICS:package_2', 'SUPPLYCHAIN:warehouse'].map((brand) => {
            const [name, icon] = brand.split(':');
            return (
              <div key={name} className="flex items-center gap-2 font-black text-xl text-slate-800 dark:text-white font-display">
                <span className="material-symbols-outlined text-3xl">{icon}</span> {name}
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
};

export default TrustedBy;
