locals {
    base_string = "green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats green day sum41 all american rejects blink182 feeder the offspring the killers the strokes the white stripes the hives the vines the libertines the bravery the kooks the fratellis the wombats"
    repeat_count = 1024
}

pipeline "big_data" {

    step "transform" "big_0" {
        value = "small"
    }

    step "transform" "big_1" {
        depends_on = [step.transform.big_0]
        value = "small"
    }

    step "transform" "big_2" {
        depends_on = [step.transform.big_1]
        value = join("", [for i in range(local.repeat_count) : local.base_string])
    }

    output "val" {
        value = step.transform.big_2.value
    }
}