const ro = {
  translation: {
    // Navigation
    nav: {
      dashboard: 'Panou',
      targets: 'Conturi',
      newActivity: 'Activitate nouă',
      planner: 'Planificator',
      mapPlanner: 'Planificator hartă',
      team: 'Echipă',
      settings: 'Setări',
      help: 'Ajutor',
      switchAccount: 'Schimbă contul',
    },

    // Theme
    theme: {
      label: 'Temă',
      light: 'Deschisă',
      dark: 'Întunecată',
    },

    // Language
    language: {
      label: 'Limbă',
      en: 'English',
      ro: 'Română',
    },

    // Common
    common: {
      loading: 'Se încarcă...',
      save: 'Salvează',
      cancel: 'Anulează',
      confirm: 'Confirmă',
      goBack: 'Înapoi',
      search: 'Caută',
      export: 'Exportă CSV',
      today: 'Astăzi',
      page: 'Pagina',
      of: 'din',
      results: 'rezultate',
      noResults: '0 rezultate',
      select: '— Selectează —',
      all: 'Toate',
    },

    // Dashboard
    dashboard: {
      title: 'Centru de comandă',
      subtitle: 'KPI-uri bazate pe activități pentru echipa ta',
      loading: 'Se încarcă panoul...',
      completionRate: 'Rată de finalizare',
      planned: 'Planificate',
      completed: 'Finalizate',
    },

    // Dashboard cards
    dashboardCards: {
      activities: 'Activități',
      total: 'total',
      fieldVsNonField: 'Teren vs Birou',
      field: 'Teren',
      nonField: 'Birou',
      targetCoverage: 'Acoperire conturi',
      frequencyCompliance: 'Conformitate frecvență',
      visits: 'vizite',
      targets_count: '{{count}} conturi, {{required}} necesare/cont',
      teamPerformance: 'Performanță echipă',
      weekly: 'SĂPTĂMÂNAL',
      monthly: 'LUNAR',
    },

    // Targets
    targets: {
      title: 'Conturi',
      subtitle: 'Gestionarea a {{count}} conturi din toate regiunile.',
      allTypes: 'Toate tipurile',
      totalTargets: 'Total conturi',
      name: 'Nume',
      type: 'Tip',
      location: 'Locație',
      frequency: 'Frecvență',
      noTargets: 'Nu s-au găsit conturi.',
      noTargetsOfType: 'Nu există conturi de tipul „{{type}}".',
      loading: 'Se încarcă conturile...',
      backToTargets: 'Înapoi la conturi',
      details: 'Detalii',
      activitiesSection: 'Activități',
      activitiesPlaceholder: 'Urmărirea activităților va fi disponibilă după implementarea funcționalității de activități.',
      loadingTarget: 'Se încarcă contul...',
      failedToLoad: 'Nu s-a putut încărca contul.',
      notFound: 'Contul nu a fost găsit.',
    },

    // Planner
    planner: {
      title: 'Planificator',
      subtitle: 'Planifică și urmărește activitățile de teren.',
      week: 'Săptămână',
      month: 'Lună',
      myActivities: 'Activitățile mele',
      statusLegend: 'Legendă status',
      categories: 'Categorii',
      fieldActivities: 'Activități de teren',
      nonFieldActivities: 'Activități de birou',
      dailyPulse: 'Puls zilnic',
      todayCount: '{{count}} Astăzi',
      inView: '{{count}} Vizibile',
      loading: 'Se încarcă planificatorul...',
      recoveryDays: 'Zile de recuperare',
      recoveryAvailable: '{{count}} disponibile',
      recoveryEarnedTaken: '{{earned}} acumulate, {{taken}} luate',
      claimBy: 'Revendică până la',
    },

    // Daily view
    daily: {
      backToPlanner: 'Înapoi la planificator',
      noActivities: 'Nu sunt activități programate pentru această zi.',
      createActivity: 'Creează o activitate',
      activitiesScheduled_one: '{{count}} activitate programată',
      activitiesScheduled_other: '{{count}} activități programate',
      submitted: 'Trimis',
    },

    // Activities
    activity: {
      newTitle: 'Activitate nouă',
      editTitle: 'Editare activitate',
      create: 'Creează activitatea',
      update: 'Actualizează activitatea',
      saving: 'Se salvează...',
      type: 'Tip activitate',
      status: 'Status',
      date: 'Dată',
      duration: 'Durată',
      target: 'Cont',
      selectType: '— Selectează tipul —',
      selectStatus: '— Selectează statusul —',
      selectDuration: '— Selectează durata —',
      searchTargets: 'Caută conturi...',
      searchTargetsEllipsis: 'Caută conturi\u2026',
      selected: 'Selectat: {{name}}',
      details: 'Detalii',
      fields: 'Câmpuri {{type}}',
      detailsLabel: 'Detalii {{type}}',
      loading: 'Se încarcă activitatea...',
      loadingConfig: 'Se încarcă configurația...',
      failedToLoad: 'Nu s-a putut încărca activitatea.',
      notFound: 'Activitatea nu a fost găsită.',
      failedToCreate: 'Nu s-a putut crea activitatea',
    },

    // Activity list
    activityList: {
      allStatuses: 'Toate statusurile',
      allTypes: 'Toate tipurile',
      activity_one: '{{count}} activitate',
      activity_other: '{{count}} activități',
      noActivities: 'Nu s-au găsit activități.',
      activityCol: 'Activitate',
      durationCol: 'Durată',
      statusCol: 'Status',
      loadingActivities: 'Se încarcă activitățile...',
    },

    // Activity detail
    activityDetail: {
      backToDashboard: 'Înapoi la panou',
      backToPlanner: 'Înapoi la planificator',
      backToDaily: 'Înapoi la vizualizarea zilnică',
      backToMap: 'Înapoi la planificatorul hartă',
      submitted: 'Trimis',
      jointVisit: 'Vizită comună cu {{name}}',
      markAs: 'Marchează ca {{status}}?',
      markAsMessage: 'Statusul va fi schimbat în {{status}}. Această acțiune nu poate fi anulată.',
      submitReport: 'Trimite raportul',
      submitReportQuestion: 'Trimite raportul?',
      submitReportMessage: 'Odată trimis, activitatea va fi blocată și nu mai poate fi editată.',
      submitting: 'Se trimite\u2026',
      retrySave: 'Reîncearcă salvarea',
      setStatusFirst: 'Setează statusul la finalizat sau anulat înainte de a trimite.',
      failedToSave: 'Nu s-au putut salva modificările. Apasă pentru a reîncerca.',
    },

    // Save states
    saveState: {
      saving: 'Se salvează\u2026',
      saved: 'Salvat',
      notSaved: 'Nesalvat \u2014 apasă pentru a reîncerca',
    },

    // Team
    team: {
      title: 'Management echipă',
      subtitleWithCount: 'Gestionarea a {{count}} membri ai echipei',
      subtitleEmpty: 'Prezentare generală a echipei de vânzări',
      addMember: 'Adaugă membru',
      loading: 'Se încarcă membrii echipei...',
      error: 'Nu s-au putut încărca membrii echipei. Vă rugăm să încercați din nou.',
      empty: 'Nu s-au găsit membri ai echipei.',
      assigned: 'Atribuite',
      completed: 'Finalizate',
      efficiency: 'Eficiență',
      completionRate: 'Rată de finalizare',
    },

    // Demo
    demo: {
      chooseAccount: 'Alege un cont pentru a explora demo-ul',
      demoEnvironment: 'Acesta este un mediu demo cu date de exemplu.',
      failedToSignIn: 'Autentificarea a eșuat',
      admin: 'Admin',
      manager: 'Manager',
      rep: 'Reprezentant',
    },

    // Search
    search: {
      placeholder: 'Caută conturi sau membri ai echipei...',
    },

    // Pagination
    pagination: {
      previousMonth: 'Luna anterioară',
      nextMonth: 'Luna următoare',
      previousPeriod: 'Perioada anterioară',
      nextPeriod: 'Perioada următoare',
      previousDay: 'Ziua anterioară',
      nextDay: 'Ziua următoare',
      previousPage: 'Pagina anterioară',
      nextPage: 'Pagina următoare',
    },

    // Errors
    error: {
      failedToLoadTargets: 'Nu s-au putut încărca conturile. Vă rugăm să încercați din nou.',
    },

    // Weekdays
    weekdays: {
      mon: 'Lun',
      tue: 'Mar',
      wed: 'Mie',
      thu: 'Joi',
      fri: 'Vin',
      sat: 'Sâm',
      sun: 'Dum',
    },

    weekdaysFull: {
      sunday: 'Duminică',
      monday: 'Luni',
      tuesday: 'Marți',
      wednesday: 'Miercuri',
      thursday: 'Joi',
      friday: 'Vineri',
      saturday: 'Sâmbătă',
    },
  },
} as const

export default ro
