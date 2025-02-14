<template>
	<div
		class="energyflow cursor-pointer position-relative"
		:class="{ 'energyflow--open': detailsOpen }"
		@click="toggleDetails"
	>
		<div class="row">
			<Visualization
				class="col-12 mb-3 mb-md-4"
				:gridImport="gridImport"
				:selfConsumption="selfConsumption"
				:loadpoints="loadpointsPower"
				:pvExport="pvExport"
				:batteryCharge="batteryCharge"
				:batteryDischarge="batteryDischarge"
				:pvProduction="pvProduction"
				:homePower="homePower"
				:batterySoC="batterySoC"
				:valuesInKw="valuesInKw"
			/>
		</div>
		<div class="indicator position-absolute bottom-0 start-50">
			<shopicon-regular-arrowdown></shopicon-regular-arrowdown>
		</div>
		<div class="details" :style="{ height: detailsHeight }">
			<div ref="detailsInner" class="details-inner row">
				<div class="col-12 d-flex justify-content-between pt-2 mb-4">
					<div class="d-flex flex-nowrap align-items-center">
						<span class="color-self me-2"
							><shopicon-filled-square></shopicon-filled-square
						></span>
						<span>{{ $t("main.energyflow.selfConsumption") }}</span>
					</div>
					<div v-if="gridImport > 0" class="d-flex flex-nowrap align-items-center">
						<span>{{ $t("main.energyflow.gridImport") }}</span>
						<span class="color-grid ms-2"
							><shopicon-filled-square></shopicon-filled-square
						></span>
					</div>
					<div v-else class="d-flex flex-nowrap align-items-center">
						<span>{{ $t("main.energyflow.pvExport") }}</span>
						<span class="color-export ms-2"
							><shopicon-filled-square></shopicon-filled-square
						></span>
					</div>
				</div>
				<div
					class="col-12 col-md-6 pe-md-5 pb-4 d-flex flex-column justify-content-between"
				>
					<div class="d-flex justify-content-between align-items-end mb-4">
						<h3 class="m-0">In</h3>
						<span class="fw-bold">{{ kw(inPower) }}</span>
					</div>
					<div>
						<EnergyflowEntry
							:name="$t('main.energyflow.pvProduction')"
							icon="sun"
							:power="pvProduction"
							:valuesInKw="valuesInKw"
						/>
						<EnergyflowEntry
							v-if="batteryConfigured"
							:name="$t('main.energyflow.batteryDischarge')"
							icon="battery"
							:soc="batterySoC"
							:power="batteryDischarge"
							:valuesInKw="valuesInKw"
						/>
						<EnergyflowEntry
							:name="$t('main.energyflow.gridImport')"
							icon="powersupply"
							:power="gridImport"
							:valuesInKw="valuesInKw"
						/>
					</div>
				</div>
				<div
					class="col-12 col-md-6 ps-md-5 pb-4 d-flex flex-column justify-content-between"
				>
					<div class="d-flex justify-content-between align-items-end mb-4">
						<h3 class="m-0">Out</h3>
						<span class="fw-bold">{{ kw(outPower) }}</span>
					</div>
					<div>
						<EnergyflowEntry
							:name="$t('main.energyflow.homePower')"
							icon="home"
							:power="homePower"
							:valuesInKw="valuesInKw"
						/>
						<EnergyflowEntry
							:name="
								$tc('main.energyflow.loadpoints', activeLoadpointsCount, {
									count: activeLoadpointsCount,
								})
							"
							icon="car3"
							:power="loadpointsPower"
							:valuesInKw="valuesInKw"
						/>
						<EnergyflowEntry
							v-if="batteryConfigured"
							:name="$t('main.energyflow.batteryCharge')"
							icon="battery"
							:soc="batterySoC"
							:power="batteryCharge"
							:valuesInKw="valuesInKw"
						/>
						<EnergyflowEntry
							:name="$t('main.energyflow.pvExport')"
							icon="powersupply"
							:power="pvExport"
							:valuesInKw="valuesInKw"
						/>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import "@h2d2/shopicons/es/filled/square";
import "@h2d2/shopicons/es/regular/arrowdown";
import Visualization from "./Visualization.vue";
import EnergyflowEntry from "./EnergyflowEntry.vue";
import formatter from "../../mixins/formatter";

export default {
	name: "Energyflow",
	components: { Visualization, EnergyflowEntry },
	mixins: [formatter],
	props: {
		gridConfigured: Boolean,
		gridPower: { type: Number, default: 0 },
		homePower: { type: Number, default: 0 },
		pvConfigured: Boolean,
		pvPower: { type: Number, default: 0 },
		loadpointsPower: { type: Number, default: 0 },
		activeLoadpointsCount: { type: Number, default: 0 },
		batteryConfigured: Boolean,
		batteryPower: { type: Number, default: 0 },
		batterySoC: { type: Number, default: 0 },
	},
	data: () => {
		return { detailsOpen: false, detailsCompleteHeight: null };
	},
	computed: {
		gridImport: function () {
			return Math.max(0, this.gridPower);
		},
		pvProduction: function () {
			return this.pvConfigured ? Math.abs(this.pvPower) : this.pvExport;
		},
		batteryPowerAdjusted: function () {
			const batteryPowerThreshold = 50;
			return Math.abs(this.batteryPower) < batteryPowerThreshold ? 0 : this.batteryPower;
		},
		batteryDischarge: function () {
			return Math.abs(Math.max(0, this.batteryPowerAdjusted));
		},
		batteryCharge: function () {
			return Math.abs(Math.min(0, this.batteryPowerAdjusted) * -1);
		},
		selfConsumption: function () {
			const ownPower = this.batteryDischarge + this.pvProduction;
			const consumption = this.homePower + this.batteryCharge + this.loadpointsPower;
			return Math.min(ownPower, consumption);
		},
		pvExport: function () {
			return Math.max(0, this.gridPower * -1);
		},
		valuesInKw: function () {
			return this.gridImport + this.selfConsumption + this.pvExport > 1000;
		},
		inPower: function () {
			return this.gridImport + this.pvProduction + this.batteryDischarge;
		},
		outPower: function () {
			return this.homePower + this.loadpointsPower + this.pvExport + this.batteryCharge;
		},
		detailsHeight: function () {
			return this.detailsOpen ? this.detailsCompleteHeight + "px" : 0;
		},
	},
	mounted() {
		window.addEventListener("resize", this.updateHeight);
	},
	unmounted() {
		window.removeEventListener("resize", this.updateHeight);
	},
	methods: {
		kw: function (watt) {
			return this.fmtKw(watt, this.valuesInKw);
		},
		toggleDetails: function () {
			this.updateHeight();
			this.detailsOpen = !this.detailsOpen;
		},
		updateHeight: function () {
			this.detailsCompleteHeight = this.$refs.detailsInner.offsetHeight;
		},
	},
};
</script>
<style scoped>
.energyflow {
	background: var(--bs-white);
}
.indicator {
	opacity: 0;
	transform: translateX(-50%) scaleY(1);
	transition: opacity, transform;
	transition-duration: var(--evcc-transition-slow);
}
.energyflow--open .indicator {
	transform: translateX(-50%) scaleY(-1);
}
@media (hover: hover) and (pointer: fine) {
	.energyflow:hover .indicator {
		opacity: 0.25;
	}
}
.details {
	height: 0;
	overflow: visible;
	transition: height;
	transition-duration: var(--evcc-transition-medium);
	transition-timing-function: cubic-bezier(0.5, 0.5, 0.5, 1.15);
}
.color-grid {
	color: var(--evcc-grid);
}
.color-self {
	color: var(--evcc-self);
}
.color-export {
	color: var(--evcc-export);
}
</style>
