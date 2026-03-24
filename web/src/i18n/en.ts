const en = {
  translation: {
    // Navigation
    nav: {
      dashboard: 'Dashboard',
      targets: 'Targets',
      newActivity: 'New Activity',
      planner: 'Planner',
      mapPlanner: 'Map Planner',
      team: 'Team',
      settings: 'Settings',
      help: 'Help',
      switchAccount: 'Switch Account',
    },

    // Theme
    theme: {
      label: 'Theme',
      light: 'Light',
      dark: 'Dark',
    },

    // Language
    language: {
      label: 'Language',
      en: 'English',
      ro: 'Română',
    },

    // Common
    common: {
      loading: 'Loading...',
      save: 'Save',
      cancel: 'Cancel',
      confirm: 'Confirm',
      goBack: 'Go back',
      search: 'Search',
      export: 'Export CSV',
      today: 'Today',
      page: 'Page',
      of: 'of',
      results: 'results',
      noResults: '0 results',
      select: '— Select —',
      all: 'All',
    },

    // Dashboard
    dashboard: {
      title: 'Command Center',
      subtitle: 'Activity-based KPIs for your team',
      loading: 'Loading dashboard...',
      completionRate: 'Completion Rate',
      planned: 'Planned',
      completed: 'Completed',
    },

    // Dashboard cards
    dashboardCards: {
      activities: 'Activities',
      total: 'total',
      fieldVsNonField: 'Field vs Non-field',
      field: 'Field',
      nonField: 'Non-field',
      targetCoverage: 'Target Coverage',
      frequencyCompliance: 'Frequency Compliance',
      visits: 'visits',
      targets_count: '{{count}} targets, {{required}} required/target',
      teamPerformance: 'Team Performance',
      weekly: 'WEEKLY',
      monthly: 'MONTHLY',
    },

    // Targets
    targets: {
      title: 'Targets',
      subtitle: 'Managing {{count}} targets across all regions.',
      allTypes: 'All types',
      totalTargets: 'Total Targets',
      name: 'Name',
      type: 'Type',
      location: 'Location',
      frequency: 'Frequency',
      noTargets: 'No targets found.',
      noTargetsOfType: 'No targets of type "{{type}}".',
      loading: 'Loading targets...',
      backToTargets: 'Back to targets',
      details: 'Details',
      activitiesSection: 'Activities',
      activitiesPlaceholder: 'Activity tracking will be available after the activities feature is implemented.',
      loadingTarget: 'Loading target...',
      failedToLoad: 'Failed to load target.',
      notFound: 'Target not found.',
    },

    // Planner
    planner: {
      title: 'Planner',
      subtitle: 'Plan and track field activities.',
      week: 'Week',
      month: 'Month',
      myActivities: 'My Activities',
      statusLegend: 'Status Legend',
      categories: 'Categories',
      fieldActivities: 'Field activities',
      nonFieldActivities: 'Non-field activities',
      dailyPulse: 'Daily Pulse',
      todayCount: '{{count}} Today',
      inView: '{{count}} In View',
      loading: 'Loading planner...',
      recoveryDays: 'Recovery Days',
      recoveryAvailable: '{{count}} available',
      recoveryEarnedTaken: '{{earned}} earned, {{taken}} taken',
      claimBy: 'Claim by',
    },

    // Daily view
    daily: {
      backToPlanner: 'Back to planner',
      noActivities: 'No activities scheduled for this day.',
      createActivity: 'Create an activity',
      activitiesScheduled_one: '{{count}} activity scheduled',
      activitiesScheduled_other: '{{count}} activities scheduled',
      submitted: 'Submitted',
    },

    // Activities
    activity: {
      newTitle: 'New Activity',
      editTitle: 'Edit Activity',
      create: 'Create Activity',
      update: 'Update Activity',
      saving: 'Saving...',
      type: 'Activity type',
      status: 'Status',
      date: 'Date',
      duration: 'Duration',
      target: 'Target',
      selectType: '— Select type —',
      selectStatus: '— Select status —',
      selectDuration: '— Select duration —',
      searchTargets: 'Search targets...',
      searchTargetsEllipsis: 'Search targets\u2026',
      selected: 'Selected: {{name}}',
      details: 'Details',
      fields: '{{type}} Fields',
      detailsLabel: '{{type}} Details',
      loading: 'Loading activity...',
      loadingConfig: 'Loading configuration...',
      failedToLoad: 'Failed to load activity.',
      notFound: 'Activity not found.',
      failedToCreate: 'Failed to create activity',
    },

    // Activity list
    activityList: {
      allStatuses: 'All statuses',
      allTypes: 'All types',
      activity_one: '{{count}} activity',
      activity_other: '{{count}} activities',
      noActivities: 'No activities found.',
      activityCol: 'Activity',
      durationCol: 'Duration',
      statusCol: 'Status',
      loadingActivities: 'Loading activities...',
    },

    // Activity detail
    activityDetail: {
      backToDashboard: 'Back to dashboard',
      backToPlanner: 'Back to planner',
      backToDaily: 'Back to daily view',
      backToMap: 'Back to map planner',
      submitted: 'Submitted',
      jointVisit: 'Joint visit with {{name}}',
      markAs: 'Mark as {{status}}?',
      markAsMessage: 'This will change the status to {{status}}. This action cannot be undone.',
      submitReport: 'Submit Report',
      submitReportQuestion: 'Submit report?',
      submitReportMessage: 'Once submitted, the activity will be locked and can no longer be edited.',
      submitting: 'Submitting\u2026',
      retrySave: 'Retry Save',
      setStatusFirst: 'Set status to completed or cancelled before submitting.',
      failedToSave: 'Failed to save changes. Tap retry to try again.',
    },

    // Save states
    saveState: {
      saving: 'Saving\u2026',
      saved: 'Saved',
      notSaved: 'Not saved \u2014 tap to retry',
    },

    // Team
    team: {
      title: 'Team Management',
      subtitleWithCount: 'Managing {{count}} team members',
      subtitleEmpty: 'Overview of your sales team',
      addMember: 'Add Member',
      loading: 'Loading team members...',
      error: 'Failed to load team members. Please try again.',
      empty: 'No team members found.',
      assigned: 'Assigned',
      completed: 'Completed',
      efficiency: 'Efficiency',
      completionRate: 'Completion rate',
    },

    // Demo
    demo: {
      chooseAccount: 'Choose an account to explore the demo',
      demoEnvironment: 'This is a demo environment with sample data.',
      failedToSignIn: 'Failed to sign in',
      admin: 'Admin',
      manager: 'Manager',
      rep: 'Rep',
    },

    // Search
    search: {
      placeholder: 'Search targets or team members...',
    },

    // Pagination
    pagination: {
      previousMonth: 'Previous month',
      nextMonth: 'Next month',
      previousPeriod: 'Previous period',
      nextPeriod: 'Next period',
      previousDay: 'Previous day',
      nextDay: 'Next day',
      previousPage: 'Previous page',
      nextPage: 'Next page',
    },

    // Config-driven labels (activity types, statuses, durations, options)
    configLabels: {
      // Activity types
      'type.visit': 'Visit',
      'type.administrative': 'Administrative',
      'type.business_travel': 'Business Travel',
      'type.company_event': 'Company Event',
      'type.cycle_meeting': 'Cycle Meeting',
      'type.team_meeting': 'Team Meeting',
      'type.training': 'Training',
      'type.public_holiday': 'Public Holiday',
      'type.vacation': 'Vacation',
      'type.lunch_break': 'Lunch Break',
      'type.recovery': 'Recovery Day',
      // Statuses
      'status.planned': 'Planned',
      'status.completed': 'Completed',
      'status.cancelled': 'Cancelled',
      // Durations
      'duration.full_day': 'Full Day',
      'duration.half_day': 'Half Day',
      // Visit types
      'option.visit_types.f2f': 'F2F',
      'option.visit_types.remote': 'Remote',
      // Routing options
      'option.routing_options.week_1': 'Week 1',
      'option.routing_options.week_2': 'Week 2',
      'option.routing_options.week_3': 'Week 3',
      // Account types
      'accountType.doctor': 'Doctor',
      'accountType.pharmacy': 'Pharmacy',
      // Field labels (config-driven)
      'field.name': 'Name',
      'field.specialty': 'Specialty',
      'field.potential': 'Potential',
      'field.city': 'City',
      'field.county': 'County',
      'field.address': 'Address',
      'field.pharmacy_type': 'Pharmacy Type',
      'field.account_id': 'Account',
      'field.visit_type': 'Visit Type',
      'field.promoted_products': 'Promoted Products',
      'field.feedback': 'Feedback',
      'field.details': 'Details',
      'field.joint_visit_user_id': 'Joint Visit User',
      'field.routing': 'Routing',
      'field.duration': 'Duration',
    },

    // Errors
    error: {
      failedToLoadTargets: 'Failed to load targets. Please try again.',
    },

    // Weekdays
    weekdays: {
      mon: 'Mon',
      tue: 'Tue',
      wed: 'Wed',
      thu: 'Thu',
      fri: 'Fri',
      sat: 'Sat',
      sun: 'Sun',
    },

    weekdaysFull: {
      sunday: 'Sunday',
      monday: 'Monday',
      tuesday: 'Tuesday',
      wednesday: 'Wednesday',
      thursday: 'Thursday',
      friday: 'Friday',
      saturday: 'Saturday',
    },
  },
} as const

export default en
